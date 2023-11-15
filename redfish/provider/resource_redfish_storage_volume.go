package provider

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &RedfishStorageVolumeResource{}
)

const (
	defaultStorageVolumeResetTimeout  int64 = 120
	defaultStorageVolumeJobTimeout    int64 = 1200
	intervalStorageVolumeJobCheckTime int64 = 10
	maxCapacityBytes                  int64 = 1000000000
)

// NewRedfishStorageVolumeResource is a helper function to simplify the provider implementation.
func NewRedfishStorageVolumeResource() resource.Resource {
	return &RedfishStorageVolumeResource{}
}

// RedfishStorageVolumeResource is the resource implementation.
type RedfishStorageVolumeResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *RedfishStorageVolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*RedfishStorageVolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "storage_volume"
}

// VolumeSchema defines the schema for the storage volume resource.
func VolumeSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"capacity_bytes": schema.Int64Attribute{
			MarkdownDescription: "Capacity Bytes",
			Description:         "Capacity Bytes",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.AtLeast(maxCapacityBytes),
			},
		},
		"disk_cache_policy": schema.StringAttribute{
			MarkdownDescription: "Disk Cache Policy",
			Description:         "Disk Cache Policy",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("Enabled"),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					"Enabled",
					"Disabled",
				}...),
			},
		},
		"drives": schema.ListAttribute{
			MarkdownDescription: "Drives",
			Description:         "Drives",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage volume resource",
			Description:         "ID of the storage volume resource",
			Computed:            true,
		},
		"optimum_io_size_bytes": schema.Int64Attribute{
			MarkdownDescription: "Optimum Io Size Bytes",
			Description:         "Optimum Io Size Bytes",
			Optional:            true,
		},
		"read_cache_policy": schema.StringAttribute{
			MarkdownDescription: "Read Cache Policy",
			Description:         "Read Cache Policy",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(string(redfish.OffReadCachePolicyType)),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.ReadAheadReadCachePolicyType),
					string(redfish.AdaptiveReadAheadReadCachePolicyType),
					string(redfish.OffReadCachePolicyType),
				}...),
			},
		},
		"reset_timeout": schema.Int64Attribute{
			MarkdownDescription: "Reset Timeout",
			Description:         "Reset Timeout",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(defaultStorageVolumeResetTimeout),
		},
		"reset_type": schema.StringAttribute{
			MarkdownDescription: "Reset Type",
			Description:         "Reset Type",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(string(redfish.ForceRestartResetType)),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.ForceRestartResetType),
					string(redfish.GracefulRestartResetType),
					string(redfish.PowerCycleResetType),
				}...),
			},
		},
		"settings_apply_time": schema.StringAttribute{
			MarkdownDescription: "Settings Apply Time",
			Description:         "Settings Apply Time",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(string(redfishcommon.ImmediateApplyTime)),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfishcommon.ImmediateApplyTime),
					string(redfishcommon.OnResetApplyTime),
				}...),
			},
		},
		"storage_controller_id": schema.StringAttribute{
			MarkdownDescription: "Storage Controller ID",
			Description:         "Storage Controller ID",
			Required:            true,
		},
		"volume_job_timeout": schema.Int64Attribute{
			MarkdownDescription: "Volume Job Timeout",
			Description:         "Volume Job Timeout",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(defaultStorageVolumeJobTimeout),
		},
		"volume_name": schema.StringAttribute{
			MarkdownDescription: "Volume Name",
			Description:         "Volume Name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"volume_type": schema.StringAttribute{
			MarkdownDescription: "Volume Type",
			Description:         "Volume Type",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.NonRedundantVolumeType),
					string(redfish.MirroredVolumeType),
					string(redfish.StripedWithParityVolumeType),
					string(redfish.SpannedMirrorsVolumeType),
					string(redfish.SpannedStripesWithParityVolumeType),
				}...),
			},
		},
		"write_cache_policy": schema.StringAttribute{
			MarkdownDescription: "Write Cache Policy",
			Description:         "Write Cache Policy",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(string(redfish.UnprotectedWriteBackWriteCachePolicyType)),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.WriteThroughWriteCachePolicyType),
					string(redfish.ProtectedWriteBackWriteCachePolicyType),
					string(redfish.UnprotectedWriteBackWriteCachePolicyType),
				}...),
			},
		},
	}
}

// Schema defines the schema for the resource.
func (*RedfishStorageVolumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing storage volume.",
		Attributes:          VolumeSchema(),
		Blocks:              RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *RedfishStorageVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_RedfishStorageVolume create : Started")
	// Get Plan Data
	var plan models.RedfishStorageVolume
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	diags = createRedfishStorageVolume(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_RedfishStorageVolume create: updating state finished, saving ...")
	// Save into State
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_RedfishStorageVolume create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *RedfishStorageVolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_RedfishStorageVolume read: started")
	var state models.RedfishStorageVolume
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	diags, cleanup := readRedfishStorageVolume(service, &state)
	if cleanup {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_RedfishStorageVolume read: finished reading state")
	// Save into State
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_RedfishStorageVolume read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *RedfishStorageVolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_RedfishStorageVolume update: started")
	var state, plan models.RedfishStorageVolume
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan Data
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	diags = updateRedfishStorageVolume(ctx, service, &plan, &state)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_RedfishStorageVolume update: finished state update")
	// Save into State
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_RedfishStorageVolume update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *RedfishStorageVolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_RedfishStorageVolume delete: started")
	// Get State Data
	var state models.RedfishStorageVolume
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	diags = deleteRedfishStorageVolume(ctx, service, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_RedfishStorageVolume delete: finished")
}

func createRedfishStorageVolume(ctx context.Context, service *gofish.Service, d *models.RedfishStorageVolume) diag.Diagnostics {
	var diags diag.Diagnostics
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(d.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(d.RedfishServer[0].Endpoint.ValueString())

	// Get user config
	storageID := d.StorageControllerID.ValueString()
	volumeType := d.VolumeType.ValueString()
	volumeName := d.VolumeName.ValueString()
	optimumIOSizeBytes := int(d.OptimumIoSizeBytes.ValueInt64())
	capacityBytes := int(d.CapacityBytes.ValueInt64())
	readCachePolicy := d.ReadCachePolicy.ValueString()
	writeCachePolicy := d.WriteCachePolicy.ValueString()
	diskCachePolicy := d.DiskCachePolicy.ValueString()
	applyTime := d.SettingsApplyTime.ValueString()
	volumeJobTimeout := int64(d.VolumeJobTimeout.ValueInt64())

	var driveNames []string
	diags.Append(d.Drives.ElementsAs(ctx, &driveNames, true)...)

	// Get storage
	storage, err := getStorage(service, storageID)
	if err != nil {
		diags.AddError("Error when retreiving the Storage from the Redfish API", err.Error())
		return diags
	}

	// Check if settings_apply_time is doable on this controller
	err = checkSettingsApplyTime(storage, applyTime)
	if err != nil {
		diags.AddError("Error while checking support for settings_apply_time", err.Error())
		return diags
	}

	// Get drives
	allStorageDrives, err := storage.Drives()
	if err != nil {
		diags.AddError("Error when getting the drives attached to controller", err.Error())
		return diags
	}
	drives, err := getDrives(allStorageDrives, driveNames)
	if err != nil {
		diags.AddError("Error when getting the drives", err.Error())
		return diags
	}

	newVolume := map[string]interface{}{
		"VolumeType":         volumeType,
		"DisplayName":        volumeName,
		"Name":               volumeName,
		"ReadCachePolicy":    readCachePolicy,
		"WriteCachePolicy":   writeCachePolicy,
		"CapacityBytes":      capacityBytes,
		"OptimumIOSizeBytes": optimumIOSizeBytes,
		"Oem": map[string]map[string]map[string]interface{}{
			"Dell": {
				"DellVolume": {
					"DiskCachePolicy": diskCachePolicy,
				},
			},
		},
		"@Redfish.OperationApplyTime": applyTime,
	}

	var listDrives []map[string]string
	for _, drive := range drives {
		storageDrive := make(map[string]string)
		storageDrive["@odata.id"] = drive.Entity.ODataID
		listDrives = append(listDrives, storageDrive)
	}
	newVolume["Drives"] = listDrives

	// Create volume job
	jobID, err := createVolume(service, storage.ODataID, newVolume)
	if err != nil {
		diags.AddError("Error when creating the virtual disk on disk controller", err.Error())
		return diags
	}

	// Immediate or OnReset scenarios
	if applyTime == string(redfishcommon.OnResetApplyTime) { // OnReset case
		// Get reset_timeout and reset_type from schema
		resetType := d.ResetType.ValueString()
		resetTimeout := d.ResetTimeout.ValueInt64()

		// Reboot the server
		pOp := powerOperator{ctx, service}
		_, err := pOp.PowerOperation(resetType, resetTimeout, intervalStorageVolumeJobCheckTime)
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}

	// Wait for the job to finish
	err = common.WaitForJobToFinish(service, jobID, intervalStorageVolumeJobCheckTime, volumeJobTimeout)
	if err != nil {
		diags.AddError(RedfishJobErrorMsg, err.Error())
		return diags
	}

	// Get storage volumes
	volumes, err := storage.Volumes()
	if err != nil {
		diags.AddError("there was an issue when retrieving volumes", err.Error())
		return diags
	}
	volumeID, err := getVolumeID(volumes, volumeName)
	if err != nil {
		diags.AddError("Error. The volume ID with given volume name was not found", err.Error())
		return diags
	}

	d.ID = types.StringValue(volumeID)
	return diags
}

func readRedfishStorageVolume(service *gofish.Service, d *models.RedfishStorageVolume) (diags diag.Diagnostics, cleanup bool) {
	// var diags diag.Diagnostics

	// Check if the volume exists
	_, err := redfish.GetVolume(service.GetClient(), d.ID.ValueString())
	if err != nil {
		e, ok := err.(*redfishcommon.Error)
		if !ok {
			diags.AddError("There was an error with the API", err.Error())
			return diags, false
		}
		if e.HTTPReturnedStatusCode == http.StatusNotFound {
			diags.AddError("Volume doesn't exist", "")
			return diags, true
		}
		diags.AddError("Status code", err.Error())
		return diags, false
	}

	/*
		- If it has jobID, if finished, get the volumeID
		Also never EVER trigger an update regarding disk properties for safety reasons
	*/

	return diags, false
}

func updateRedfishStorageVolume(ctx context.Context, service *gofish.Service,
	d *models.RedfishStorageVolume, state *models.RedfishStorageVolume,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(d.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(d.RedfishServer[0].Endpoint.ValueString())

	// Get user config
	storageID := d.StorageControllerID.ValueString()
	volumeName := d.VolumeName.ValueString()
	readCachePolicy := d.ReadCachePolicy.ValueString()
	writeCachePolicy := d.WriteCachePolicy.ValueString()
	diskCachePolicy := d.DiskCachePolicy.ValueString()
	applyTime := d.SettingsApplyTime.ValueString()

	var driveNames []string
	diags.Append(d.Drives.ElementsAs(ctx, &driveNames, true)...)

	volumeJobTimeout := d.ResetTimeout.ValueInt64()

	// Get storage
	storage, err := getStorage(service, storageID)
	if err != nil {
		diags.AddError("Error when retreiving storage details from the Redfish API", err.Error())
		return diags
	}

	// Check if settings_apply_time is doable on this controller
	err = checkSettingsApplyTime(storage, applyTime)
	if err != nil {
		diags.AddError("Error while checking support for settings_apply_time", err.Error())
		return diags
	}

	payload := map[string]interface{}{
		"ReadCachePolicy":  readCachePolicy,
		"WriteCachePolicy": writeCachePolicy,
		"DisplayName":      volumeName,
		"Oem": map[string]map[string]map[string]interface{}{
			"Dell": {
				"DellVolume": {
					"DiskCachePolicy": diskCachePolicy,
				},
			},
		},
		"Name": volumeName,
		"@Redfish.SettingsApplyTime": map[string]interface{}{
			"ApplyTime": applyTime,
		},
	}

	// Update volume job
	jobID, err := updateVolume(service, state.ID.ValueString(), payload)
	if err != nil {
		diags.AddError("Error when updating the virtual disk on disk controller", err.Error())
		return diags
	}

	// Immediate or OnReset scenarios
	if applyTime == string(redfishcommon.OnResetApplyTime) { // OnReset case
		resetType := d.ResetType.ValueString()
		resetTimeout := d.ResetTimeout.ValueInt64()

		// Reboot the server
		pOp := powerOperator{ctx, service}
		_, err := pOp.PowerOperation(resetType, resetTimeout, intervalStorageVolumeJobCheckTime)
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}

	// Wait for the job to finish
	err = common.WaitForJobToFinish(service, jobID, intervalStorageVolumeJobCheckTime, volumeJobTimeout)
	if err != nil {
		diags.AddError(RedfishJobErrorMsg, err.Error())
		return diags
	}

	// Get storage volumes
	volumes, err := storage.Volumes()
	if err != nil {
		diags.AddError("there was an issue when retrieving volumes", err.Error())
		return diags
	}
	volumeID, err := getVolumeID(volumes, volumeName)
	if err != nil {
		diags.AddError("Error. The volume ID with given volume name was not found", err.Error())
		return diags
	}

	d.ID = types.StringValue(volumeID)
	return diags
}

func deleteRedfishStorageVolume(ctx context.Context, service *gofish.Service, d *models.RedfishStorageVolume) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(d.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(d.RedfishServer[0].Endpoint.ValueString())

	// Get vars from schema
	applyTime := d.SettingsApplyTime.ValueString()
	volumeJobTimeout := d.VolumeJobTimeout.ValueInt64()

	jobID, err := deleteVolume(service, d.ID.ValueString())
	if err != nil {
		diags.AddError("Error. There was an error when deleting volume", err.Error())
		return diags
	}

	if applyTime == string(redfishcommon.OnResetApplyTime) { // OnReset case
		// Get reset_timeout and reset_type from schema
		resetType := d.ResetType.ValueString()
		resetTimeout := d.ResetTimeout.ValueInt64()

		// Reboot the server
		pOp := powerOperator{ctx, service}
		_, err := pOp.PowerOperation(resetType, resetTimeout, intervalStorageVolumeJobCheckTime)
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}

	// WAIT FOR VOLUME TO DELETE
	err = common.WaitForJobToFinish(service, jobID, intervalStorageVolumeJobCheckTime, volumeJobTimeout)
	if err != nil {
		diags.AddError("Error, timeout reached when waiting for job to finish", err.Error())
		return diags
	}

	return diags
}

func getStorageController(storageControllers []*redfish.Storage, diskControllerID string) (*redfish.Storage, error) {
	for _, storage := range storageControllers {
		if storage.Entity.ID == diskControllerID {
			return storage, nil
		}
	}
	return nil, fmt.Errorf("error. Didn't find the storage controller %v", diskControllerID)
}

func deleteVolume(service *gofish.Service, volumeURI string) (jobID string, err error) {
	// TODO - Check if we can delete immediately or if we need to schedule a job
	res, err := service.GetClient().Delete(volumeURI)
	if err != nil {
		return "", fmt.Errorf("error while deleting the volume %s", volumeURI)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("the operation was not successful. Return code was different from 202 ACCEPTED")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("there was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getDrives(drives []*redfish.Drive, driveNames []string) ([]*redfish.Drive, error) {
	drivesToReturn := []*redfish.Drive{}
	for _, v := range drives {
		for _, w := range driveNames {
			if v.Name == w {
				drivesToReturn = append(drivesToReturn, v)
			}
		}
	}
	if len(driveNames) != len(drivesToReturn) {
		return nil, fmt.Errorf("any of the drives you inserted doesn't exist")
	}
	return drivesToReturn, nil
}

func checkSettingsApplyTime(storage *redfish.Storage, applyTime string) error {
	operationApplyTimes, err := storage.GetOperationApplyTimeValues()
	if err != nil {
		return fmt.Errorf("couldn't retrieve operationApplyTimes from controller: %w", err)
	}
	if !checkOperationApplyTimes(applyTime, operationApplyTimes) {
		return fmt.Errorf("Storage controller does not support settings_apply_time: %s", applyTime)
	}
	return nil
}

func getStorage(service *gofish.Service, storageID string) (*redfish.Storage, error) {
	systems, err := service.Systems()
	if err != nil {
		return nil, fmt.Errorf("Error when retreiving the Systems from the Redfish API: %w", err)
	}

	storageControllers, err := systems[0].Storage()
	if err != nil {
		return nil, fmt.Errorf("Error when retreiving the Storage from the Redfish API: %w", err)
	}

	storage, err := getStorageController(storageControllers, storageID)
	if err != nil {
		return nil, fmt.Errorf("Error when getting the storage struct: %w", err)
	}
	return storage, nil
}

/*
createVolume creates a virtualdisk on a disk controller by using the redfish API
*/
func createVolume(service *gofish.Service,
	storageLink string,
	newVolume map[string]interface{},
) (jobID string, err error) {
	volumesURL := fmt.Sprintf("%v/Volumes", storageLink)

	res, err := service.GetClient().Post(volumesURL, newVolume)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("the query was unsucessfull")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("there was some error when retreiving the jobID")
	}
	return jobID, nil
}

func updateVolume(service *gofish.Service,
	storageLink string,
	payload map[string]interface{},
) (jobID string, err error) {
	volumesURL := fmt.Sprintf("%v/Settings", storageLink)

	res, err := service.GetClient().Patch(volumesURL, payload)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("the query was unsucessfull")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("there was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getVolumeID(volumes []*redfish.Volume, volumeName string) (volumeLink string, err error) {
	for _, v := range volumes {
		if v.Name == volumeName {
			volumeLink = v.ODataID
			return volumeLink, nil
		}
	}
	return "", fmt.Errorf("couldn't find a volume with the provided name: %s", volumeName)
}

func checkOperationApplyTimes(optionToCheck string, storageOperationApplyTimes []redfishcommon.OperationApplyTime) (result bool) {
	for _, v := range storageOperationApplyTimes {
		if optionToCheck == string(v) {
			return true
		}
	}
	return false
}

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/dell/terraform-provider-redfish/common"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
// 	"github.com/stmcginnis/gofish"
// 	redfishcommon "github.com/stmcginnis/gofish/common"
// 	"github.com/stmcginnis/gofish/redfish"
// )

// const (
// 	defaultStorageVolumeResetTimeout  int = 120
// 	defaultStorageVolumeJobTimeout    int = 1200
// 	intervalStorageVolumeJobCheckTime int = 10
// )

// func resourceRedfishStorageVolume() *schema.Resource {
// 	return &schema.Resource{
// 		CreateContext: resourceRedfishStorageVolumeCreate,
// 		ReadContext:   resourceRedfishStorageVolumeRead,
// 		UpdateContext: resourceRedfishStorageVolumeUpdate,
// 		DeleteContext: resourceRedfishStorageVolumeDelete,
// 		Schema:        getResourceRedfishStorageVolumeSchema(),
// 	}
// }

// func getResourceRedfishStorageVolumeSchema() map[string]*schema.Schema {
// 	return map[string]*schema.Schema{
// 		"redfish_server": {
// 			Type:        schema.TypeList,
// 			Required:    true,
// 			Description: "This list contains the different redfish endpoints to manage (different servers)",
// 			Elem: &schema.Resource{
// 				Schema: map[string]*schema.Schema{
// 					"user": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "This field is the user to login against the redfish API",
// 					},
// 					"password": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "This field is the password related to the user given",
// 					},
// 					"endpoint": {
// 						Type:        schema.TypeString,
// 						Required:    true,
// 						Description: "This field is the endpoint where the redfish API is placed",
// 					},
// 					"ssl_insecure": {
// 						Type:        schema.TypeBool,
// 						Optional:    true,
// 						Description: "This field indicates if the SSL/TLS certificate must be verified",
// 					},
// 				},
// 			},
// 		},
// 		"storage_controller_id": {
// 			Type:        schema.TypeString,
// 			Required:    true,
// 			Description: "This value must be the storage controller ID the user want to manage. I.e: RAID.Integrated.1-1",
// 		},
// 		"volume_name": {
// 			Type:         schema.TypeString,
// 			Required:     true,
// 			Description:  "This value is the desired name for the volume to be given",
// 			ValidateFunc: validation.StringLenBetween(1, 15),
// 		},
// 		"volume_type": {
// 			Type:        schema.TypeString,
// 			Required:    true,
// 			Description: "This value specifies the raid level the virtual disk is going to have.",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.NonRedundantVolumeType),
// 				string(redfish.MirroredVolumeType),
// 				string(redfish.StripedWithParityVolumeType),
// 				string(redfish.SpannedMirrorsVolumeType),
// 				string(redfish.SpannedStripesWithParityVolumeType),
// 			}, false),
// 		},
// 		"drives": {
// 			Type:        schema.TypeList,
// 			Required:    true,
// 			Description: "This list contains the physical disks names to create the volume within a disk controller",
// 			Elem: &schema.Schema{
// 				Type: schema.TypeString,
// 			},
// 		},
// 		"settings_apply_time": {
// 			Type:        schema.TypeString,
// 			Description: "Flag to make the operation either \"Immediate\" or \"OnReset\". By default value is \"Immediate\"",
// 			Optional:    true,
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfishcommon.ImmediateApplyTime),
// 				string(redfishcommon.OnResetApplyTime)}, false),
// 			Default: string(redfishcommon.ImmediateApplyTime),
// 		},
// 		"reset_type": {
// 			Type:     schema.TypeString,
// 			Optional: true,
// 			Description: "Reset type allows to choose the type of restart to apply when settings_apply_time is set to \"OnReset\"" +
// 				"Possible values are: \"ForceRestart\", \"GracefulRestart\" or \"PowerCycle\". If not set, \"ForceRestart\" is the default.",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.ForceRestartResetType),
// 				string(redfish.GracefulRestartResetType),
// 				string(redfish.PowerCycleResetType),
// 			}, false),
// 			Default: string(redfish.ForceRestartResetType),
// 		},
// 		"reset_timeout": {
// 			Type:     schema.TypeInt,
// 			Optional: true,
// 			Description: "reset_timeout is the time in seconds that the provider waits for the server to be reset" +
// 				"(if settings_apply_time is set to \"OnReset\") before timing out. Default is 120s.",
// 			Default: defaultStorageVolumeResetTimeout,
// 		},
// 		"volume_job_timeout": {
// 			Type:     schema.TypeInt,
// 			Optional: true,
// 			Description: "volume_job_timeout is the time in seconds that the provider waits for the volume job" +
// 				"Default is 1200s",
// 			Default: defaultStorageVolumeJobTimeout,
// 		},
// 		"capacity_bytes": {
// 			Type:         schema.TypeInt,
// 			Optional:     true,
// 			Description:  "capacity_bytes shall contain the size in bytes of the associated volume.",
// 			ValidateFunc: validation.IntAtLeast(1000000000),
// 		},
// 		"optimum_io_size_bytes": {
// 			Type:        schema.TypeInt,
// 			Optional:    true,
// 			Description: "optimum_io_size_bytes shall contain the optimum IO size to use when performing IO on this volume.",
// 		},
// 		"read_cache_policy": {
// 			Type:        schema.TypeString,
// 			Optional:    true,
// 			Description: "read_cache_policy shall contain a boolean indicator of the read cache policy for the Volume.",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.ReadAheadReadCachePolicyType),
// 				string(redfish.AdaptiveReadAheadReadCachePolicyType),
// 				string(redfish.OffReadCachePolicyType),
// 			}, false),
// 			Default: string(redfish.OffReadCachePolicyType),
// 		},
// 		"write_cache_policy": {
// 			Type:        schema.TypeString,
// 			Optional:    true,
// 			Description: "write_cache_policy shall contain a boolean indicator of the write cache policy for the Volume.",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.WriteThroughWriteCachePolicyType),
// 				string(redfish.ProtectedWriteBackWriteCachePolicyType),
// 				string(redfish.UnprotectedWriteBackWriteCachePolicyType),
// 			}, false),
// 			Default: string(redfish.UnprotectedWriteBackWriteCachePolicyType),
// 		},
// 		"disk_cache_policy": {
// 			Type:        schema.TypeString,
// 			Optional:    true,
// 			Description: "disk_cache_policy shall contain a boolean indicator of the disk cache policy for the Volume.",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				"Enabled",
// 				"Disabled",
// 			}, false),
// 			Default: "Enabled",
// 		},
// 	}
// }

// func resourceRedfishStorageVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	return createRedfishStorageVolume(service, d)
// }

// func resourceRedfishStorageVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	return readRedfishStorageVolume(service, d)
// }

// func resourceRedfishStorageVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if diags := updateRedfishStorageVolume(ctx, service, d, m); diags.HasError() {
// 		return diags
// 	}
// 	return resourceRedfishStorageVolumeRead(ctx, d, m)
// }

// func resourceRedfishStorageVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	return deleteRedfishStorageVolume(service, d)
// }
