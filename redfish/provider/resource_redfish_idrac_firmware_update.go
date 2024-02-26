package provider

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	installURLLink    = "/Oem/Dell/DellSoftwareInstallationService/Actions/DellSoftwareInstallationService.InstallFromRepository"
	getUpdatesURLLink = "/Oem/Dell/DellSoftwareInstallationService/Actions/DellSoftwareInstallationService.GetRepoBasedUpdateList"
)

const (
	timeBetweenAttemptsCatalogUpdate = 20
	timeoutForCatalogUpdate          = 720
	defaultPort                      = 80
	httpString                       = "HTTP"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &idracFirmwareUpdateResource{}
)

// NewIdracFirmwareUpdateResource is a helper function to simplify the provider implementation.
func NewIdracFirmwareUpdateResource() resource.Resource {
	return &idracFirmwareUpdateResource{}
}

// idracFirmwareUpdateResource is the resource implementation.
type idracFirmwareUpdateResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *idracFirmwareUpdateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*idracFirmwareUpdateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "idrac_firmware_update"
}

func idracFirmwareUpdateSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:         "ID of the iDRAC Firmware Update Resource.",
			MarkdownDescription: "ID of the iDRAC Firmware Update Resource.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"share_type": schema.StringAttribute{
			Description:         "Type of the Network Share.",
			MarkdownDescription: "Type of the Network Share.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("CIFS", "NFS", "HTTP", "HTTPS", "FTP", "TFTP"),
			},
		},
		"ip_address": schema.StringAttribute{
			Description:         "IP address for the remote share.",
			MarkdownDescription: "IP address for the remote share.",
			Required:            true,
		},
		"share_name": schema.StringAttribute{
			Description: "Name of the CIFS share or full path to the NFS share. Optional for HTTP/HTTPS share (if supported)," +
				"this may be treated as the path of the directory containing the file.",
			MarkdownDescription: "Name of the CIFS share or full path to the NFS share. Optional for HTTP/HTTPS share (if supported)" +
				"this may be treated as the path of the directory containing the file.",
			Optional: true,
		},
		"catalog_file_name": schema.StringAttribute{
			Description: "Name of the catalog file on the repository. Default is Catalog.xml.",
			Computed:    true,
			Optional:    true,
			Default:     stringdefault.StaticString("Catalog.xml"),
		},
		"ignore_cert_warning": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("On"),
			Description: "Specifies if certificate warning should be ignored when HTTPS is used. If ignore_cert_warning is On," +
				"warnings are ignored. Default is On.",
			MarkdownDescription: "Specifies if certificate warning should be ignored when HTTPS is used. If ignore_cert_warning is On," +
				"warnings are ignored. Default is On.",
		},
		"share_user": schema.StringAttribute{
			Description: "Network share user in the format 'user@domain' or 'domain\\user' if user is part of a domain else 'user'." +
				"This option is mandatory for CIFS Network Share.",
			MarkdownDescription: "Network share user in the format 'user@domain' or 'domain\\user' if user is part of a domain else 'user'." +
				"This option is mandatory for CIFS Network Share.",
			Optional: true,
		},
		"share_password": schema.StringAttribute{
			Description:         "Network share user password. This option is mandatory for CIFS Network Share.",
			MarkdownDescription: "Network share user password. This option is mandatory for CIFS Network Share.",
			Optional:            true,
		},
		"proxy_support": schema.StringAttribute{
			Description:         "Specifies if a proxy should be used. Default is Off. This option is only used for HTTP, HTTPS, and FTP shares.",
			MarkdownDescription: "Specifies if a proxy should be used. Default is Off. This option is only used for HTTP, HTTPS, and FTP shares.",
			Computed:            true,
			Optional:            true,
			Default:             stringdefault.StaticString("Off"),
			Validators: []validator.String{
				stringvalidator.OneOf("ParametersProxy", "Off"),
			},
		},
		"proxy_server": schema.StringAttribute{
			Description: "The IP address of the proxy server.This IP will not be validated. The download job will be created even for" +
				"invalid proxy_server.Please check the results of the job for error details.This is required when proxy_support is ParametersProxy.",
			MarkdownDescription: "The IP address of the proxy server.This IP will not be validated. The download job will be created even for" +
				"invalid proxy_server.Please check the results of the job for error details.This is required when proxy_support is ParametersProxy.",
			Optional: true,
		},
		"proxy_port": schema.Int64Attribute{
			Description:         "The Port for the proxy server.Default is set to 80.",
			MarkdownDescription: "The Port for the proxy server.Default is set to 80.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(defaultPort),
		},
		"proxy_type": schema.StringAttribute{
			Description:         "The proxy type of the proxy server. Default is (HTTP).",
			MarkdownDescription: "The proxy type of the proxy server. Default is (HTTP).",
			Computed:            true,
			Optional:            true,
			Default:             stringdefault.StaticString(httpString),
			Validators: []validator.String{
				stringvalidator.OneOf("HTTP", "SOCKS"),
			},
		},
		"proxy_username": schema.StringAttribute{
			Description:         "The user name for the proxy server.",
			MarkdownDescription: "The user name for the proxy server.",
			Optional:            true,
		},
		"proxy_password": schema.StringAttribute{
			Description:         "The password for the proxy server.",
			MarkdownDescription: "The password for the proxy server.",
			Optional:            true,
		},
		"mount_point": schema.StringAttribute{
			Description:         "The local directory where the share should be mounted.",
			MarkdownDescription: "The local directory where the share should be mounted.",
			Optional:            true,
		},
		"apply_update": schema.BoolAttribute{
			Description: "If ApplyUpdate is set to true, the updatable packages from Catalog XML are staged. If it is set to False," +
				"no updates are applied but the list of updatable packages can be seen in the UpdateList.Default is true.",
			MarkdownDescription: "If ApplyUpdate is set to true, the updatable packages from Catalog XML are staged. If it is set to False, " +
				"no updates are applied but the list of updatable packages can be seen in the UpdateList.Default is true.",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(true),
		},
		"reboot_needed": schema.BoolAttribute{
			Description: "This property indicates if a reboot should be performed. True indicates that the system (host) is rebooted during" +
				"the update process. False indicates that the updates take effect after the system is rebooted the next time.Default is true.",
			MarkdownDescription: "This property indicates if a reboot should be performed. True indicates that the system (host) is rebooted during" +
				"the update process. False indicates that the updates take effect after the system is rebooted the next time.Default is true.",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(true),
		},
		"update_list": schema.ListNestedAttribute{
			Description:         "List of properties of the update list.",
			MarkdownDescription: "List of properties of the update list.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"package_name": schema.StringAttribute{
						Description:         "Name of the package to be updated.",
						MarkdownDescription: "Name of the package to be updated.",
						Computed:            true,
					},
					"current_package_version": schema.StringAttribute{
						Description:         "Current version of the package.",
						MarkdownDescription: "Current version of the package.",
						Computed:            true,
					},
					"target_package_version": schema.StringAttribute{
						Description:         "Target version of the package.",
						MarkdownDescription: "Target version of the package.",
						Computed:            true,
					},
					"criticality": schema.StringAttribute{
						Description:         "Criticality of the package update.",
						MarkdownDescription: "Criticality of the package update.",
						Computed:            true,
					},
					"reboot_type": schema.StringAttribute{
						Description:         "Reboot type of the package update.",
						MarkdownDescription: "Reboot type of the package update.",
						Computed:            true,
					},
					"display_name": schema.StringAttribute{
						Description:         "Display name of the package.",
						MarkdownDescription: "Display name of the package.",
						Computed:            true,
					},
					"job_id": schema.StringAttribute{
						Description:         "ID of the job if it's triggered.",
						MarkdownDescription: "ID of the job if it's triggered.",
						Computed:            true,
					},
					"job_status": schema.StringAttribute{
						Description:         "Status of the job if it's triggered.",
						MarkdownDescription: "Status of the job if it's triggered.",
						Computed:            true,
					},
					"job_message": schema.StringAttribute{
						Description:         "Message from the job if it's triggered.",
						MarkdownDescription: "Message from the job if it's triggered.",
						Computed:            true,
					},
				},
			},
		},
	}
}

// Schema defines the schema for the resource.
func (*idracFirmwareUpdateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to Update firmware of the iDRAC Server based on a catalog.",
		Description:         "This Terraform resource is used to Update firmware of the iDRAC Server based on a catalog.",
		Attributes:          idracFirmwareUpdateSchema(),
		Blocks:              RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *idracFirmwareUpdateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "redfish_idrac_firmware_update create : Started")
	// Get Plan Data
	var plan models.IdracFirmwareUpdate
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

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
	systemId := system.ODataID
	plan.Id = types.StringValue("idrac_firmware_update")

	payload, payloadError := GetInstallFirmwareUpdatePayload(plan)

	if payloadError != nil {
		resp.Diagnostics.AddError("Payload error", payloadError.Error())
		return
	}

	res, err := service.GetClient().Post(fmt.Sprintf("%v%v", systemId, installURLLink), payload)
	if err != nil {
		resp.Diagnostics.AddError("Post Install error", err.Error())
		return
	}
	if res.StatusCode != http.StatusAccepted {
		resp.Diagnostics.AddError("Post Install error", "the query was unsucessfull")
		return
	}
	repoUpdateJobId := ExtractJobID(res.Header.Get("Location"))
	if repoUpdateJobId == "" {
		resp.Diagnostics.AddError("Check repository Updates job error", "job id not found")
		return
	}
	repoUpdateJob, err := common.GetJobDetailsOnFinish(service, fmt.Sprintf("/redfish/v1/JobService/Jobs/%s", repoUpdateJobId),
		int64(common.TimeBetweenAttempts), int64(common.Timeout))
	if err != nil {
		resp.Diagnostics.AddError("Check repository Updates job error", err.Error())
		return
	}
	if repoUpdateJob != nil {
		if repoUpdateJob.JobStatus == "Critical" {
			resp.Diagnostics.AddError("Check repository Updates job error", repoUpdateJob.Messages[0].Message)
			return
		}
	}

	getpayload := map[string]interface{}{}
	getres, err := service.GetClient().Post(fmt.Sprintf("%v%v", systemId, getUpdatesURLLink), getpayload)
	if err != nil {
		resp.Diagnostics.AddError("install service error", err.Error())
		return
	}
	defer getres.Body.Close()

	getBody, err := io.ReadAll(getres.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading response body", err.Error())
		return
	}

	var response models.GetPackageListResponse
	if err := json.Unmarshal([]byte(getBody), &response); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling the pacakageList", err.Error())
		return
	}

	result, err := ParseXML(response.PackageList)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing PackageList XML:", err.Error())
		return
	}

	var jobErrors []string
	var jobs []redfish.Job
	if plan.ApplyUpdate.ValueBool() {
		jobErrors, jobs = TrackJobs(ctx, result, service)
		if len(jobErrors) > 0 {
			combinedErrorMessage := strings.Join(jobErrors, ", ")
			resp.Diagnostics.AddError("One or more jobs failed:", combinedErrorMessage)
			return
		}
	}

	// Use input values from plan
	var state models.IdracFirmwareUpdate = plan
	// Update the list of updates available using the parsed response
	state.UpdateList, _ = GetUpdatedList(result, jobs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the resource and writes to state
func (*idracFirmwareUpdateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "redfish_idrac_firmware_update read : Started")
	// Get Plan Data
	var state models.IdracFirmwareUpdate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource
func (*idracFirmwareUpdateResource) Update(ctx context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Firmware updates are done through create.
	tflog.Trace(ctx, "redfish_idrac_firmware_update update : Started")
}

// Delete removes resource from state
func (*idracFirmwareUpdateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.IdracFirmwareUpdate

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}

// TrackJobs reads the parsed XML data object and returns the list of job errors if any and job details
func TrackJobs(ctx context.Context, result []map[string]interface{}, service *gofish.Service) ([]string, []redfish.Job) {
	var jobErrors []string
	var jobs []redfish.Job

	for _, data := range result {
		properties := data["properties"].([]map[string]interface{})
		for _, prop := range properties {
			if prop["name"] == "JobID" {
				tflog.Info(ctx, "JobID matched:"+prop["value"].(string))
				jobID := prop["value"].(string)
				if jobID != "" {
					jobURI := fmt.Sprintf("/redfish/v1/JobService/Jobs/%s", jobID)
					job, err := common.GetJobDetailsOnFinish(service, jobURI, timeBetweenAttemptsCatalogUpdate, timeoutForCatalogUpdate)
					if err != nil {
						jobErrors = append(jobErrors, fmt.Sprintf("Job %s failed: %v", jobID, err.Error()))
					}
					if job != nil {
						jobs = append(jobs, *job)
					}
				}
			}
		}
	}
	if len(jobs) == 0 {
		jobErrors = append(jobErrors, "Job details not found.")
	}
	return jobErrors, jobs
}

// GetUpdatedList reads the parsed XML data object and returns the list of updates along with job information if available
func GetUpdatedList(updateListData []map[string]interface{}, jobs []redfish.Job) (types.List, diag.Diagnostics) {
	updateKey := map[string]attr.Type{
		"package_name":            types.StringType,
		"current_package_version": types.StringType,
		"target_package_version":  types.StringType,
		"criticality":             types.StringType,
		"reboot_type":             types.StringType,
		"display_name":            types.StringType,
		"job_id":                  types.StringType,
		"job_status":              types.StringType,
		"job_message":             types.StringType,
	}
	var updateObjects []attr.Value
	for _, data := range updateListData {
		properties := data["properties"].([]map[string]interface{})
		// Create a map for properties
		updateMap := make(map[string]attr.Value)
		for _, prop := range properties {
			jobId := ""
			propertyName := prop["name"].(string)
			var propertyValue string
			if propValue, ok := prop["value"].(string); ok {
				propertyValue = propValue
			} else {
				propertyValue = ""
			}

			switch propertyName {
			case "PackageName":
				updateMap["package_name"] = types.StringValue(propertyValue)
			case "PackageVersion":
				updateMap["target_package_version"] = types.StringValue(propertyValue)
			case "ComponentInstalledVersion":
				updateMap["current_package_version"] = types.StringValue(propertyValue)
			case "Criticality":
				updateMap["criticality"] = types.StringValue(propertyValue)
			case "RebootType":
				updateMap["reboot_type"] = types.StringValue(propertyValue)
			case "DisplayName":
				updateMap["display_name"] = types.StringValue(propertyValue)
			case "JobID":
				updateMap["job_id"] = types.StringValue(propertyValue)
				jobId = propertyValue
			default:
			}
			const jobStatus = "job_status"
			const jobMessage = "job_message"
			if len(jobs) > 0 {
				for _, job := range jobs {
					if job.ID == jobId {
						updateMap[jobStatus] = types.StringValue(string(job.JobState))
						updateMap[jobMessage] = types.StringValue(job.Messages[0].Message)
					}
				}
			} else {
				// if user just wants to get the updates and not actually apply them, these will be empty
				updateMap[jobStatus] = types.StringValue("")
				updateMap[jobMessage] = types.StringValue("")
			}
		}
		updateObject, _ := types.ObjectValue(updateKey, updateMap)
		updateObjects = append(updateObjects, updateObject)
	}
	return types.ListValue(types.ObjectType{AttrTypes: updateKey}, updateObjects)
}

// ExtractJobID extracts the job ID from the URL
func ExtractJobID(url string) string {
	// Split the URL by "/"
	parts := strings.Split(url, "/")

	for i, part := range parts {
		if part == "Jobs" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// GetInstallFirmwareUpdatePayload returns the payload for the install firmware update request
func GetInstallFirmwareUpdatePayload(plan models.IdracFirmwareUpdate) (map[string]interface{}, error) {
	// create payload with required fields
	payload := map[string]interface{}{
		"RebootNeeded": plan.RebootNeeded.ValueBool(),
		"IPAddress":    plan.IPAddress.ValueString(),
		"ShareType":    plan.ShareType.ValueString(),
	}
	if plan.ApplyUpdate.ValueBool() {
		payload["ApplyUpdate"] = "True"
	} else {
		payload["ApplyUpdate"] = "False"
	}
	// update payload with optional fields
	if plan.ShareUser.ValueString() != "" && plan.SharePassword.ValueString() != "" {
		payload["UserName"] = plan.ShareUser.ValueString()
		payload["Password"] = plan.SharePassword.ValueString()
	} else if plan.ShareUser.ValueString() == "" || plan.SharePassword.ValueString() == "" {
		if plan.ShareType.ValueString() == "CIFS" {
			return nil, fmt.Errorf("ShareUser and SharePassword are required when ShareType is CIFS")
		}
	}
	// add sharename if sharetype is nfs or cifs
	if plan.ShareName.ValueString() != "" {
		payload["ShareName"] = plan.ShareName.ValueString()
	} else {
		if plan.ShareType.ValueString() == "CIFS" || plan.ShareType.ValueString() == "NFS" {
			return nil, fmt.Errorf("ShareName is required when ShareType is CIFS or NFS")
		}
	}
	// Add proxy details if proxy support is enabled
	if plan.ProxySupport.ValueString() != "Off" {
		if plan.ShareType.ValueString() != httpString && plan.ShareType.ValueString() != "HTTPS" && plan.ShareType.ValueString() != "FTP" {
			return nil, fmt.Errorf("Proxy is only supported when ShareType is HTTP, HTTPS or FTP")
		}
		if plan.ProxyServer.ValueString() == "" {
			return nil, fmt.Errorf("ProxyServer is required when ProxySupport is Enabled")
		}
		payload["ProxySupport"] = plan.ProxySupport.ValueString()
		payload["ProxyServer"] = plan.ProxyServer.ValueString()
		payload["ProxyPort"] = plan.ProxyPort.ValueInt64()
		if plan.ProxyUsername.ValueString() != "" && plan.ProxyPassword.ValueString() != "" {
			payload["ProxyUname"] = plan.ProxyUsername.ValueString()
			payload["ProxyPasswd"] = plan.ProxyPassword.ValueString()
		}
	}
	if plan.CatalogFileName.ValueString() != "" {
		payload["CatalogFile"] = plan.CatalogFileName.ValueString()
	}
	if plan.MountPoint.ValueString() != "" {
		payload["Mountpoint"] = plan.MountPoint.ValueString()
	}

	payload["IgnoreCertWarning"] = plan.IgnoreCertificateWarning.ValueString()
	return payload, nil
}

// ParseXML parses the XML and returns the map of updates
func ParseXML(xmlData string) ([]map[string]interface{}, error) {
	var cim models.CIM
	if err := xml.NewDecoder(strings.NewReader(xmlData)).Decode(&cim); err != nil {
		return nil, err
	}
	const name = "name"
	const value = "value"
	updateList := make([]map[string]interface{}, 0)
	for _, instance := range cim.Instances {
		update := make(map[string]interface{})
		update["class_name"] = instance.ClassName
		properties := make([]map[string]interface{}, 0)
		for _, prop := range instance.Properties {
			propertyName := prop.Name
			propertyValue := prop.Value
			propertyMap := map[string]interface{}{
				name:  propertyName,
				value: propertyValue,
			}
			properties = append(properties, propertyMap)
		}

		if len(instance.PropertyArrays) > 0 {
			for _, propArray := range instance.PropertyArrays {
				propertyName := propArray.Name
				for _, value := range propArray.Values {
					propertyMap := map[string]interface{}{
						name:  propertyName,
						value: value,
					}
					properties = append(properties, propertyMap)
				}
			}
		}
		update["properties"] = properties
		updateList = append(updateList, update)
	}

	return updateList, nil
}
