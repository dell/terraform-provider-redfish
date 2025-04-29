/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	defaultSimpleUpdateResetTimeout  int   = 120
	defaultSimpleUpdateJobTimeout    int   = 1200
	intervalSimpleUpdateJobCheckTime int64 = 10
	locationKey                            = "Location"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &simpleUpdateResource{}
)

// NewSimpleUpdateResource is a helper function to simplify the provider implementation.
func NewSimpleUpdateResource() resource.Resource {
	return &simpleUpdateResource{}
}

// simpleUpdateResource is the resource implementation.
type simpleUpdateResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *simpleUpdateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*simpleUpdateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "simple_update"
}

func simpleUpdateSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:         "ID of the simple update resource",
			MarkdownDescription: "ID of the simple update resource",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
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
				"Make sure you place your firmware packages in the same folder as the module and set " +
				"it as follows: \"${path.module}/BIOS_FXC54_WN64_1.15.0.EXE\"",
			// DiffSuppressFunc will allow moving fw packages through the filesystem without triggering an update if so.
			// At the moment it uses filename to see if they're the same. We need to strengthen that by somehow using hashing
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIf(
					func(
						_ context.Context,
						req planmodifier.StringRequest,
						resp *stringplanmodifier.RequiresReplaceIfFuncResponse,
					) {
						spath, ppath := req.StateValue.ValueString(), req.ConfigValue.ValueString()
						resp.RequiresReplace = true
						if filepath.Base(spath) == filepath.Base(ppath) {
							resp.RequiresReplace = false
						}
					},
					"",
					"",
				),
			},
		},
		"reset_type": schema.StringAttribute{
			Required: true,
			Description: "Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled." +
				" Possible values are: \"ForceRestart\", \"GracefulRestart\" or \"PowerCycle\"",
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
			Description: "Time in seconds that the provider waits for the server to be reset before timing out.",
		},
		"simple_update_job_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(int64(defaultSimpleUpdateJobTimeout)),
			Description: "Time in seconds that the provider waits for the simple update job to be completed before timing out.",
		},
		"software_id": schema.StringAttribute{
			Computed:    true,
			Description: "Software ID from the firmware package uploaded",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"version": schema.StringAttribute{
			Computed:    true,
			Description: "Software version from the firmware package uploaded",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"system_id": schema.StringAttribute{
			MarkdownDescription: "System ID of the system",
			Description:         "System ID of the system",
			Computed:            true,
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// Schema defines the schema for the resource.
func (*simpleUpdateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to Update firmware of the iDRAC Server." +
			" We can Read the existing firmware version or update the same using this resource.",
		Description: "This Terraform resource is used to Update firmware of the iDRAC Server." +
			" We can Read the existing firmware version or update the same using this resource.",

		Attributes: simpleUpdateSchema(),
		Blocks: map[string]schema.Block{
			"redfish_server": schema.ListNestedBlock{
				MarkdownDescription: "List of server BMCs and their respective user credentials",
				Description:         "List of server BMCs and their respective user credentials",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: RedfishServerSchema(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
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
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()
	system, err := getSystemResource(service, plan.SystemID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("system error", err.Error())
		return
	}
	plan.SystemID = types.StringValue(system.ID)
	plan.Id = types.StringValue(system.SerialNumber + "_simple_update")

	// resetType := plan.DesiredPowerAction.ValueString()
	updater := simpleUpdater{
		ctx:     ctx,
		service: service,
	}
	dia, state := updater.updateRedfishSimpleUpdate(plan)
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
	api, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()
	dia, newState := readRedfishSimpleUpdate(service, state)
	resp.Diagnostics.Append(dia...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update also refreshes the resource and writes to state
func (*simpleUpdateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// update can be triggerred by only a change in image path, where base name of image remains same
	// So set plan to state.
	tflog.Trace(ctx, "resource_simple_update update : Started")
	// Get Plan Data
	var plan models.SimpleUpdateRes
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes resource from state
func (*simpleUpdateResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func readRedfishSimpleUpdate(service *gofish.Service, d models.SimpleUpdateRes) (diag.Diagnostics, models.SimpleUpdateRes) {
	var diags diag.Diagnostics

	// Try to get software inventory
	_, err := redfish.GetSoftwareInventory(service.GetClient(), d.Id.ValueString())
	if err != nil {
		var redfishErr *redfishcommon.Error
		if !errors.As(err, &redfishErr) {
			diags.AddError("there was an issue with the API", err.Error())
		} else {
			// the firmware package previously applied has changed, trigger update
			d.Image = types.StringNull()
		}
	}

	return diags, d
}

type simpleUpdater struct {
	ctx           context.Context
	service       *gofish.Service
	updateService *redfish.UpdateService
}

func (u *simpleUpdater) updateRedfishSimpleUpdate(d models.SimpleUpdateRes) (diag.Diagnostics, models.SimpleUpdateRes) {
	var diags diag.Diagnostics
	ret := d

	transferProtocol := d.Protocol.ValueString()
	targetFirmwareImage := d.Image.ValueString()
	resetType := d.ResetType.ValueString()

	// Check if chosen reset type is supported before doing anything else
	system, err := getSystemResource(u.service, d.SystemID.ValueString())
	if err != nil {
		diags.AddError(
			"Couldn't retrieve allowed reset types from systems",
			err.Error(),
		)
		return diags, ret
	}
	tflog.Debug(u.ctx, "resource_simple_update : found system")
	d.SystemID = types.StringValue(system.ID)

	if ok := checkResetType(resetType, system.SupportedResetTypes); !ok {
		diags.AddError(
			fmt.Sprintf("Reset type %s is not available in this redfish implementation", resetType),
			err.Error(),
		)
		return diags, ret
	}
	tflog.Debug(u.ctx, "resource_simple_update : reset type "+resetType+"is available")

	// Get update service from root
	updateService, err := u.service.UpdateService()
	if err != nil {
		diags.AddError("error while retrieving UpdateService", err.Error())
		return diags, ret
	}
	tflog.Debug(u.ctx, "resource_simple_update : found update service")
	u.updateService = updateService

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
	tflog.Debug(u.ctx, "resource_simple_update : update type "+transferProtocol+" is valid")

	if transferProtocol == "NFS" {
		tflog.Info(u.ctx, "Remote NFS protocol detected")
		ret, err = u.pullUpdate(ret)
		if err != nil {
			diags.AddError(err.Error(), "")
		}
		tflog.Debug(u.ctx, "Update Complete")
	} else if transferProtocol == "HTTP" || transferProtocol == "HTTPS" {
		if strings.HasPrefix(targetFirmwareImage, "http") {
			tflog.Info(u.ctx, "Remote HTTP protocol detected")
			ret, err = u.pullUpdate(ret)
			if err != nil {
				diags.AddError(err.Error(), "")
			}
			tflog.Debug(u.ctx, "Update Complete")
		} else {
			tflog.Info(u.ctx, "Local firmware detected")

			// 17G check
			service := u.service
			isGenerationSeventeenAndAbove, err := isServerGenerationSeventeenAndAbove(service)
			if err != nil {
				diags.AddError("Error retrieving the server generation", err.Error())
				return diags, ret
			}

			if isGenerationSeventeenAndAbove {
				ret, err = u.uploadLocalFirmwareSeventeenGeneration(ret)
				if err != nil {
					diags.AddError(err.Error(), "")
				}
				tflog.Debug(u.ctx, "Update Complete")
			} else {
				// var fwPackage *redfish.SoftwareInventory
				fwPackage, err := u.uploadLocalFirmware(ret)
				if err != nil {
					// TBD - HOW TO HANDLE WHEN FAILS BUT FIRMWARE WAS INSTALLED?
					diags.AddError(err.Error(), "")
					return diags, ret
				}
				ret.SoftwareId = types.StringValue(fwPackage.SoftwareID)
				ret.Version = types.StringValue(fwPackage.Version)
				ret.Id = types.StringValue(fwPackage.ODataID)
				tflog.Info(u.ctx, "Uploading Local Firmware Complete")

				diagsRead, state := readRedfishSimpleUpdate(u.service, ret)
				diags.Append(diagsRead...)
				ret = state
			}
		}
	} else {
		diags.AddError("Transfer protocol not available in this implementation", "")
	}

	return diags, ret
}

func (u *simpleUpdater) uploadLocalFirmware(d models.SimpleUpdateRes) (*redfish.SoftwareInventory, error) {
	// Get ETag from FW inventory
	service, updateService := u.service, u.updateService

	dellUpdateService, err := dell.UpdateService(updateService)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving FirmwareInventory URI: %w", err)
	}

	response, err := service.GetClient().Get(dellUpdateService.FirmwareInventory.ODataID)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving Etag from FirmwareInventory: %w", err)
	}
	response.Body.Close() // #nosec G104
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
	response.Body.Close() // #nosec G104
	packageLocation := response.Header.Get(locationKey)

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
	response, err = service.GetClient().Post(dellUpdateService.SimpleUpdateActions.SimpleUpdate.Target, triggerUpdatePayload)
	if err != nil {
		// Delete uploaded package - TBD
		return nil, fmt.Errorf("there was an issue when scheduling the update job - %w", err)
	}
	response.Body.Close() // #nosec G104

	// Get jobid
	jobID := response.Header.Get(locationKey)
	d.Id = types.StringValue(jobID)
	err = u.updateJobStatus(d)
	if err != nil {
		return nil, fmt.Errorf("error running job %w", err)
	}
	tflog.Debug(u.ctx, "resource_simple_update : Job finished successfully")
	// Get updated FW inventory
	// sleep time to allow the inventory service to get started
	time.Sleep(30 * time.Second)
	fwInventory, err := updateService.FirmwareInventories()
	if err != nil {
		// TBD - HOW TO HANDLE WHEN FAILS BUT FIRMWARE WAS INSTALLED?
		return nil, fmt.Errorf("error when getting firmware inventory - %w", err)
	}
	tflog.Debug(u.ctx, "resource_simple_update : Retrieved Firmware Inventories")

	inv, err := getFWfromInventory(fwInventory, packageInformation.SoftwareID, packageInformation.Version)
	if err != nil {
		err = fmt.Errorf("error when retrieving fw package from fw inventory - %w", err)
	}
	tflog.Debug(u.ctx, "resource_simple_update : Retrieved Status from Inventories")
	return inv, err
}

func (u *simpleUpdater) uploadLocalFirmwareSeventeenGeneration(d models.SimpleUpdateRes) (models.SimpleUpdateRes, error) {
	// Get update service from root
	updateService := u.updateService
	service := u.service

	customHeaders := map[string]string{}

	// Open file to upload
	file, err := openFile(d.Image.ValueString())
	if err != nil {
		return d, fmt.Errorf("couldn't open FW file to upload - %w", err)
	}
	defer file.Close()

	// Set payload
	payload := map[string]io.Reader{
		"UpdateFile": file,
	}

	// Upload FW Package to FW inventory
	response, err := service.GetClient().PostMultipartWithHeaders(updateService.MultipartHTTPPushURI, payload, customHeaders)
	if err != nil {
		return d, fmt.Errorf("there was an issue when uploading FW package to redfish - %w", err)
	}

	// Get jobid
	jobID := response.Header.Get(locationKey)
	tflog.Info(u.ctx, "resource_simple_update : Job is scheduled with id "+jobID)

	// changes for 17G - replacing TaskMonitors with Tasks
	jobID = strings.Replace(jobID, "TaskMonitors", "Tasks", 1)

	d.Id = types.StringValue(jobID)
	err = u.updateJobStatus(d)
	if err != nil {
		return d, fmt.Errorf("there was an issue when waiting for the job to complete - %w", err)
	}

	job, err := redfish.GetTask(service.GetClient(), jobID)
	if len(job.Messages) > 0 {
		message := job.Messages[0].Message
		if strings.Contains(message, "Unable to transfer") || strings.Contains(message, "Module took more time than expected.") {
			err = errors.Join(err, fmt.Errorf("please check the image path, download failed"))
		}
	}
	if err != nil {
		return d, err
	}
	tflog.Info(u.ctx, "Retrieved successful task")

	swInventory, err := redfish.GetSoftwareInventory(service.GetClient(), d.Id.ValueString())
	if err != nil {
		return d, fmt.Errorf("unable to fetch data %w", err)
	}
	tflog.Debug(u.ctx, "Retrieved inventory with ID "+swInventory.ODataID)

	d.Id = types.StringValue(swInventory.ODataID)
	d.Version = types.StringValue(swInventory.Version)
	d.SoftwareId = types.StringValue(swInventory.SoftwareID)
	return d, nil
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
	f, err := os.Open(filePath) // #nosec G304
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

func (u *simpleUpdater) pullUpdate(d models.SimpleUpdateRes) (models.SimpleUpdateRes, error) {
	// Get update service from root
	updateService := u.updateService
	service := u.service

	dellUpdateService, err := dell.UpdateService(updateService)
	if err != nil {
		return d, fmt.Errorf("error while retrieving dellUpdate service: %w", err)
	}

	protocol := d.Protocol.ValueString()
	imagePath := d.Image.ValueString()
	httpURI := dellUpdateService.SimpleUpdateActions.SimpleUpdate.Target

	payload := make(map[string]interface{})
	payload["ImageURI"] = imagePath
	payload["TransferProtocol"] = protocol
	tflog.Trace(u.ctx, fmt.Sprintf("resource_simple_update : Job is scheduling payload %v", payload))

	response, err := service.GetClient().Post(httpURI, payload)
	if err != nil {
		// Delete uploaded package - TBD
		return d, fmt.Errorf("there was an issue when scheduling the update job - %w", err)
	}

	// Get jobid
	jobID := response.Header.Get(locationKey)
	tflog.Info(u.ctx, "resource_simple_update : Job is scheduled with id "+jobID)

	// changes for 17G - replacing TaskMonitors with Tasks
	jobID = strings.Replace(jobID, "TaskMonitors", "Tasks", 1)

	d.Id = types.StringValue(jobID)
	err = u.updateJobStatus(d)
	if err != nil {
		// Delete uploaded package - TBD
		return d, fmt.Errorf("there was an issue when waiting for the job to complete - %w", err)
	}

	job, err := redfish.GetTask(service.GetClient(), jobID)
	if len(job.Messages) > 0 {
		message := job.Messages[0].Message
		if strings.Contains(message, "Unable to transfer") || strings.Contains(message, "Module took more time than expected.") {
			err = errors.Join(err, fmt.Errorf("please check the image path, download failed"))
		}
	}
	if err != nil {
		return d, err
	}
	tflog.Info(u.ctx, "Retrieved successful task")

	swInventory, err := redfish.GetSoftwareInventory(service.GetClient(), d.Id.ValueString())
	if err != nil {
		return d, fmt.Errorf("unable to fetch data %w", err)
	}
	tflog.Debug(u.ctx, "Retrieved inventory with ID "+swInventory.ODataID)

	d.Id = types.StringValue(swInventory.ODataID)
	d.Version = types.StringValue(swInventory.Version)
	d.SoftwareId = types.StringValue(swInventory.SoftwareID)
	return d, nil
}

func (u *simpleUpdater) updateJobStatus(d models.SimpleUpdateRes) error {
	// Get jobid
	jobID := d.Id.ValueString()

	resetTimeout := d.ResetTimeout.ValueInt64()
	simpleUpdateJobTimeout := d.JobTimeout.ValueInt64()
	tflog.Debug(u.ctx, fmt.Sprintf(
		"resource_simple_update : resetTimeout is set to %d and simpleUpdateJobTimeout to %d",
		resetTimeout,
		simpleUpdateJobTimeout))

	// Reboot the server
	tflog.Debug(u.ctx, "Rebooting the server")
	pOp := powerOperator{u.ctx, u.service, d.SystemID.ValueString()}
	_, err := pOp.PowerOperation(d.ResetType.ValueString(), resetTimeout, intervalSimpleUpdateJobCheckTime)
	if err != nil {
		// Delete uploaded package - TBD
		return fmt.Errorf("there was an issue when restarting the server: %w", err)
	}
	tflog.Debug(u.ctx, "Reboot Complete")

	// Check JID
	err = common.WaitForTaskToFinish(u.service, jobID, intervalSimpleUpdateJobCheckTime, simpleUpdateJobTimeout)
	if err != nil {
		// Delete uploaded package - TBD
		return fmt.Errorf("there was an issue when waiting for the job to complete - %w", err)
	}
	tflog.Debug(u.ctx, "Job has been completed")

	return nil
}
