package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultSimpleUpdateResetTimeout  int = 120
	defaultSimpleUpdateJobTimeout    int = 1200
	intervalSimpleUpdateJobCheckTime int = 10
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &simpleUpdateResource{}
)

// NewpowerResource is a helper function to simplify the provider implementation.
func NewSimpleUpdateResource() resource.Resource {
	return &simpleUpdateResource{}
}

// powerResource is the resource implementation.
type simpleUpdateResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *simpleUpdateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (r *simpleUpdateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "simple_update"
}

func SimpleUpdateSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:         "ID of the simple update resource",
			MarkdownDescription: "ID of the simple update resource",
			Computed:            true,
		},
		"redfish_server": schema.SingleNestedAttribute{
			MarkdownDescription: "Redfish Server",
			Description:         "Redfish Server",
			Required:            true,
			Attributes:          RedfishServerSchema(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
		},
		"transfer_protocol": schema.StringAttribute{
			Required: true,
			Description: "The network protocol that the Update Service uses to retrieve the software image file located at the URI provided " +
				"in ImageURI, if the URI does not contain a scheme." +
				" Accepted values: CIFS, FTP, SFTP, HTTP, HTTPS, NSF, SCP, TFTP, OEM, NFS." +
				" Currently only HTTP, HTTPS and NFS are supported with local file path or HTTP(s)/NFS link.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		/* For the time being, target_firmware_image will be the local path for our firmware packages.
		   It is intended to work along HTTP transfer protocol
		   In the future it could be used for targetting FTP/CIFS/NFS images
		   TBD - Think about a custom diff function that grabs only the file name and not the path, to avoid unneeded update triggers
		*/
		"target_firmware_image": schema.StringAttribute{
			Required: true,
			Description: "Target firmware image used for firmware update on the redfish instance. " +
				"Make sure you place your firmware packages in the same folder as the module and set it as follows: \"${path.module}/BIOS_FXC54_WN64_1.15.0.EXE\"",
			// DiffSuppressFunc will allow moving fw packages through the filesystem without triggering an update if so.
			// At the moment it uses filename to see if they're the same. We need to strengthen that by somehow using hashing
			// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			// 	if filepath.Base(old) == filepath.Base(new) {
			// 		return true
			// 	}
			// 	return false
			// },
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIf(
					func(
						_ context.Context,
						req planmodifier.StringRequest,
						resp *stringplanmodifier.RequiresReplaceIfFuncResponse,
					) {
						spath, ppath := req.StateValue.ValueString(), req.ConfigValue.ValueString()
						if filepath.Base(spath) == filepath.Base(ppath) {
							resp.RequiresReplace = false
						}
						resp.RequiresReplace = true
					},
					"",
					"",
				),
			},
		},
		"reset_type": schema.StringAttribute{
			Required: true,
			Description: "Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled." +
				"Possible values are: \"ForceRestart\", \"GracefulRestart\" or \"PowerCycle\"",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.ForceRestartResetType),
					string(redfish.GracefulRestartResetType),
					string(redfish.PowerCycleResetType),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"reset_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(int64(defaultSimpleUpdateResetTimeout)),
			Description: "reset_timeout is the time in seconds that the provider waits for the server to be reset before timing out.",
		},
		"simple_update_job_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(int64(defaultSimpleUpdateJobTimeout)),
			Description: "simple_update_job_timeout is the time in seconds that the provider waits for the simple update job to be completed before timing out.",
		},
		"software_id": schema.StringAttribute{
			Computed:    true,
			Description: "Software ID from the firmware package uploaded",
		},
		"version": schema.StringAttribute{
			Computed:    true,
			Description: "Software version from the firmware package uploaded",
		},
	}
}

// Schema defines the schema for the resource.
func (r *simpleUpdateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing power.",
		Version:             1,
		Attributes:          SimpleUpdateSchema(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *simpleUpdateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_simple_update create : Started")
	// Get Plan Data
	var plan models.SimpleUpdateRes
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	system, err := getSystemResource(service)
	if err != nil {
		resp.Diagnostics.AddError("system error", err.Error())
		return
	}

	plan.Id = types.StringValue(system.SerialNumber + "_simple_update")

	// resetType := plan.DesiredPowerAction.ValueString()
	dia, state := updateRedfishSimpleUpdate(ctx, service, plan)
	resp.Diagnostics.Append(dia...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the resource and writes to state
func (r *simpleUpdateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_simple_update read : Started")
	// Get Plan Data
	var state models.SimpleUpdateRes
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	dia, newState := readRedfishSimpleUpdate(service, state)
	resp.Diagnostics.Append(dia...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update also refreshes the resource and writes to state
func (r *simpleUpdateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// update can be triggerred by only a change in non-functional requirements.
	// So set them to state.
	tflog.Trace(ctx, "resource_simple_update update : Started")
	// Get Plan Data
	var state models.SimpleUpdateRes
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	dia, newState := readRedfishSimpleUpdate(service, state)
	resp.Diagnostics.Append(dia...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete removes resource from state
func (r *simpleUpdateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func readRedfishSimpleUpdate(service *gofish.Service, d models.SimpleUpdateRes) (diag.Diagnostics, models.SimpleUpdateRes) {
	var diags diag.Diagnostics

	// Try to get software inventory
	_, err := redfish.GetSoftwareInventory(service.GetClient(), d.Id.ValueString())
	if err != nil {
		_, ok := err.(*redfishcommon.Error)
		if !ok {
			diags.AddError("there was an issue with the API", err.Error())
		} else {
			// the firmware package previously applied has changed, trigger update
			d.Image = types.StringValue("none")
		}
	}

	return diags, d
}

func updateRedfishSimpleUpdate(ctx context.Context, service *gofish.Service, d models.SimpleUpdateRes) (diag.Diagnostics, models.SimpleUpdateRes) {
	var diags diag.Diagnostics
	ret := d

	transferProtocol := d.Protocol.ValueString()
	targetFirmwareImage := d.Image.ValueString()
	resetType := d.ResetType.ValueString()

	// Check if chosen reset type is supported before doing anything else
	systems, err := service.Systems()
	if err != nil {
		diags.AddError(
			"Couldn't retrieve allowed reset types from systems",
			err.Error(),
		)
		return diags, ret
	}

	if ok := checkResetType(resetType, systems[0].SupportedResetTypes); !ok {
		diags.AddError(
			fmt.Sprintf("Reset type %s is not available in this redfish implementation", resetType),
			err.Error(),
		)
		return diags, ret
	}

	// Get update service from root
	updateService, err := service.UpdateService()
	if err != nil {
		diags.AddError("error while retrieving UpdateService", err.Error())
		return diags, ret
	}

	// Check if the transfer protocol is available in the redfish instance
	err = checkTransferProtocol(transferProtocol, updateService)
	if err != nil {
		var availableTransferProtocols string
		for _, v := range updateService.TransferProtocol {
			availableTransferProtocols += fmt.Sprintf("%s ", v)
		}
		diags.AddError(
			err.Error(),
			fmt.Sprintf("Supported transfer protocols in this implementation: %s", availableTransferProtocols),
		)
		return diags, ret // !!!! append list of supported transfer protocols
	}

	if transferProtocol == "NFS" {
		id, err := pullUpdate(service, d)
		if err != nil {
			diags.AddError(err.Error(), "")
		} else {
			d.Id = types.StringValue(id)
		}
	} else if transferProtocol == "HTTP" || transferProtocol == "HTTPS" {
		if strings.HasPrefix(targetFirmwareImage, "http") {
			id, err := pullUpdate(service, d)
			if err != nil {
				diags.AddError(err.Error(), "")
			} else {
				d.Id = types.StringValue(id)
			}
		} else {
			fwPackage, err := uploadLocalFirmware(service, updateService, d)
			if err != nil {
				// TBD - HOW TO HANDLE WHEN FAILS BUT FIRMWARE WAS INSTALLED?
				diags.AddError(err.Error(), "")
			}
			ret.SoftwareId = types.StringValue(fwPackage.SoftwareID)
			ret.Version = types.StringValue(fwPackage.Version)
			ret.Id = types.StringValue(fwPackage.ODataID)

			diagsRead, state := readRedfishSimpleUpdate(service, d)
			diags.Append(diagsRead...)
			ret = state
		}
	} else {
		diags.AddError("Transfer protocol not available in this implementation", "")
	}

	return diags, ret
}

func uploadLocalFirmware(service *gofish.Service, updateService *redfish.UpdateService, d models.SimpleUpdateRes) (*redfish.SoftwareInventory, error) {
	// Get ETag from FW inventory
	response, err := service.GetClient().Get(updateService.FirmwareInventory)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving Etag from FirmwareInventory: %w", err)
	}
	response.Body.Close()
	etag := response.Header.Get("ETag")

	// Set custom headers
	customHeaders := map[string]string{
		"if-match": etag,
	}

	// Open file to upload
	file, err := openFile(d.Image.ValueString())
	if err != nil {
		return nil, fmt.Errorf("couldn't open FW file to upload - %w", err)
	}
	defer file.Close()

	// Set payload
	payload := map[string]io.Reader{
		"file": file,
	}

	// Upload FW Package to FW inventory
	response, err = service.GetClient().PostMultipartWithHeaders(updateService.HTTPPushURI, payload, customHeaders)
	if err != nil {
		return nil, fmt.Errorf("there was an issue when uploading FW package to redfish - %w", err)
	}
	response.Body.Close()
	packageLocation := response.Header.Get("Location")

	// Get package information ( SoftwareID - Version )
	packageInformation, err := redfish.GetSoftwareInventory(service.GetClient(), packageLocation)
	if err != nil {
		return nil, fmt.Errorf("there was an issue when retrieving uploaded package information - %w", err)
	}

	// Set payload for POST call that'll trigger the update job scheduling
	triggerUpdatePayload := struct {
		ImageURI string
	}{
		ImageURI: packageLocation,
	}
	// Do the POST call against Simple.Update service
	response, err = service.GetClient().Post(updateService.UpdateServiceTarget, triggerUpdatePayload)
	if err != nil {
		// Delete uploaded package - TBD
		return nil, fmt.Errorf("there was an issue when scheduling the update job - %w", err)
	}
	response.Body.Close()

	err = updateJobStatus(service, d, response)
	if err != nil {
		return nil, fmt.Errorf("error running job %w", err)
	}
	// Get updated FW inventory
	fwInventory, err := updateService.FirmwareInventories()
	if err != nil {
		// TBD - HOW TO HANDLE WHEN FAILS BUT FIRMWARE WAS INSTALLED?
		return nil, fmt.Errorf("error when getting firmware inventory - %w", err)
	}

	inv, err := getFWfromInventory(fwInventory, packageInformation.SoftwareID, packageInformation.Version)
	if err != nil {
		err = fmt.Errorf("error when retrieving fw package from fw inventory - %w", err)
	}
	return inv, err
}

// checkResetType check if the resetType passed is within the allowableValues slice
func checkResetType(resetType string, allowableValues []redfish.ResetType) bool {
	for _, v := range allowableValues {
		if resetType == string(v) {
			return true
		}
	}
	return false
}

// openFile is a simple function that opens a file
func openFile(filePath string) (*os.File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		err = fmt.Errorf("error when opening %s file - %w", filePath, err)
	}
	return f, err
}

// checkTransferProtocol checks if the chosen transfer protocol is available in the redfish instance
func checkTransferProtocol(transferProtocol string, updateService *redfish.UpdateService) error {
	for _, v := range updateService.TransferProtocol {
		if transferProtocol == v {
			return nil
		}
	}
	return fmt.Errorf("this transfer protocol is not available in this redfish instance")
}

// getFWfromInventory get the right SoftwareInventory struct if exists
func getFWfromInventory(softwareInventories []*redfish.SoftwareInventory, softwareID, version string) (*redfish.SoftwareInventory, error) {
	for _, v := range softwareInventories {
		if v.SoftwareID == softwareID && v.Version == version {
			return v, nil
		}
	}
	return nil, fmt.Errorf("couldn't find FW on Firmware inventory")
}

func pullUpdate(service *gofish.Service, d models.SimpleUpdateRes) (string, error) {
	// Get update service from root
	updateService, err := service.UpdateService()
	if err != nil {
		return "", fmt.Errorf("error while retrieving UpdateService - %w", err)
	}

	protocol := d.Protocol.ValueString()
	imagePath := d.Image.ValueString()
	httpURI := updateService.UpdateServiceTarget

	payload := make(map[string]interface{})
	payload["ImageURI"] = imagePath
	payload["TransferProtocol"] = protocol

	response, err := service.GetClient().Post(httpURI, payload)
	if err != nil {
		// Delete uploaded package - TBD
		return "", fmt.Errorf("there was an issue when scheduling the update job - %s", err)
	}

	// Get jobid
	jobID := response.Header.Get("Location")
	err = updateJobStatus(service, d, response)
	if err != nil {
		// Delete uploaded package - TBD
		return "", fmt.Errorf("there was an issue when waiting for the job to complete - %s", err)
	}

	job, err := redfish.GetTask(service.GetClient(), jobID)
	if len(job.Messages) > 0 {
		message := job.Messages[0].Message
		if strings.Contains(message, "Unable to transfer") || strings.Contains(message, "Module took more time than expected.") {
			err = errors.Join(err, fmt.Errorf("please check the image path, download failed"))
		}
	}
	if err != nil {
		return "", err
	}

	swInventory, err := redfish.GetSoftwareInventory(service.GetClient(), d.Id.ValueString())
	if err != nil {
		return "", fmt.Errorf("unable to fetch data %v", err)
	}
	return swInventory.ODataID, nil
}

func updateJobStatus(service *gofish.Service, d models.SimpleUpdateRes, response *http.Response) error {
	// Get jobid
	jobID := response.Header.Get("Location")

	resetTimeout := d.ResetTimeout.ValueInt64()
	simpleUpdateJobTimeout := d.JobTimeout.ValueInt64()
	log.Printf("[DEBUG] resetTimeout is set to %d and simpleUpdateJobTimeout to %d", resetTimeout, simpleUpdateJobTimeout)

	// Reboot the server
	_, diags := PowerOperation(d.ResetType.ValueString(), int(resetTimeout), intervalSimpleUpdateJobCheckTime, service)
	if diags.HasError() {
		// Delete uploaded package - TBD
		return fmt.Errorf("there was an issue when restarting the server")
	}

	// Check JID
	err := common.WaitForJobToFinish(service, jobID, intervalSimpleUpdateJobCheckTime, int(simpleUpdateJobTimeout))
	if err != nil {
		// Delete uploaded package - TBD
		return fmt.Errorf("there was an issue when waiting for the job to complete - %s", err)
	}

	return nil
}
