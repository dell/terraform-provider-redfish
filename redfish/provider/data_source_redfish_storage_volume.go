package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &StorageVolumeDataSource{}

// NewStorageVolumeDataSource creates a new storage volume data source.
func NewStorageVolumeDataSource() datasource.DataSource {
	return &StorageVolumeDataSource{}
}

// StorageVolumeDataSource defines the data source implementation.
type StorageVolumeDataSource struct {
	client *gofish.APIClient
}

// Metadata returns the data source type name.
func (d *StorageVolumeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_volume"
}

// Schema returns the data source schema.
func (d *StorageVolumeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = storageVolumeDataSourceSchema()
}

// Configure adds the provider configured client to the data source.
func (d *StorageVolumeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gofish.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gofish.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *StorageVolumeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data storageVolumeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading storage volume data source")

	// Extract filter criteria
	var filter filterModel
	resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate filter: at least one of name or ID must be provided
	if filter.Name.IsNull() && filter.ID.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid Filter",
			"At least one of 'name' or 'id' must be provided in the filter block.",
		)
		return
	}

	tflog.Debug(ctx, "Filter criteria", map[string]any{
		"name": filter.Name.ValueString(),
		"id":   filter.ID.ValueString(),
	})

	// Query volume
	volume, err := d.queryVolume(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Querying Storage Volume",
			fmt.Sprintf("Could not query storage volume: %s", err.Error()),
		)
		return
	}

	// Map volume attributes to Terraform state
	d.mapVolumeToState(ctx, volume, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// queryVolume queries the storage volume based on filter criteria.
func (d *StorageVolumeDataSource) queryVolume(ctx context.Context, filter filterModel) (*redfish.Volume, error) {
	// Get service root
	service := d.client.Service

	// Get systems
	systems, err := service.Systems()
	if err != nil {
		return nil, fmt.Errorf("failed to get systems: %w", err)
	}

	if len(systems) == 0 {
		return nil, fmt.Errorf("no systems found")
	}

	// Use first system (typically there's only one)
	system := systems[0]
	tflog.Debug(ctx, "Using system", map[string]any{"system_id": system.ID})

	// Query by ID if provided
	if !filter.ID.IsNull() && filter.ID.ValueString() != "" {
		return d.queryVolumeByID(ctx, system, filter.ID.ValueString())
	}

	// Query by name
	return d.queryVolumeByName(ctx, system, filter.Name.ValueString())
}

// queryVolumeByID queries a volume by its ID.
func (d *StorageVolumeDataSource) queryVolumeByID(ctx context.Context, system *redfish.ComputerSystem, volumeID string) (*redfish.Volume, error) {
	tflog.Debug(ctx, "Querying volume by ID", map[string]any{"volume_id": volumeID})

	// Get storage controllers
	storage, err := system.Storage()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage controllers: %w", err)
	}

	// Search all controllers for the volume
	for _, controller := range storage {
		volumes, err := controller.Volumes()
		if err != nil {
			tflog.Warn(ctx, "Failed to get volumes from controller", map[string]any{
				"controller_id": controller.ID,
				"error":         err.Error(),
			})
			continue
		}

		for _, volume := range volumes {
			if volume.ID == volumeID {
				tflog.Debug(ctx, "Found volume by ID", map[string]any{"volume_id": volumeID})
				return volume, nil
			}
		}
	}

	// Volume not found - list available volumes for debugging
	availableVolumes := d.listAvailableVolumes(ctx, system)
	return nil, fmt.Errorf("volume with ID '%s' not found\n\nAvailable volumes:\n%s", volumeID, availableVolumes)
}

// queryVolumeByName queries a volume by its name.
func (d *StorageVolumeDataSource) queryVolumeByName(ctx context.Context, system *redfish.ComputerSystem, volumeName string) (*redfish.Volume, error) {
	tflog.Debug(ctx, "Querying volume by name", map[string]any{"volume_name": volumeName})

	// Get storage controllers
	storage, err := system.Storage()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage controllers: %w", err)
	}

	var matchedVolumes []*redfish.Volume

	// Search all controllers for volumes matching the name
	for _, controller := range storage {
		volumes, err := controller.Volumes()
		if err != nil {
			tflog.Warn(ctx, "Failed to get volumes from controller", map[string]any{
				"controller_id": controller.ID,
				"error":         err.Error(),
			})
			continue
		}

		for _, volume := range volumes {
			if volume.Name == volumeName {
				matchedVolumes = append(matchedVolumes, volume)
			}
		}
	}

	// Handle results
	if len(matchedVolumes) == 0 {
		availableVolumes := d.listAvailableVolumes(ctx, system)
		return nil, fmt.Errorf("volume with name '%s' not found\n\nAvailable volumes:\n%s\n\nPlease verify the volume name and try again.", volumeName, availableVolumes)
	}

	if len(matchedVolumes) > 1 {
		var volumeList strings.Builder
		for _, vol := range matchedVolumes {
			volumeList.WriteString(fmt.Sprintf("  - %s (%s)\n", vol.Name, vol.ID))
		}
		return nil, fmt.Errorf("multiple volumes match the name '%s':\n%s\nPlease refine your filter to match exactly one volume.", volumeName, volumeList.String())
	}

	tflog.Debug(ctx, "Found volume by name", map[string]any{
		"volume_name": volumeName,
		"volume_id":   matchedVolumes[0].ID,
	})

	return matchedVolumes[0], nil
}

// listAvailableVolumes lists all available volumes for error messages.
func (d *StorageVolumeDataSource) listAvailableVolumes(ctx context.Context, system *redfish.ComputerSystem) string {
	storage, err := system.Storage()
	if err != nil {
		return "(unable to list volumes)"
	}

	var volumeList strings.Builder
	for _, controller := range storage {
		volumes, err := controller.Volumes()
		if err != nil {
			continue
		}

		for _, volume := range volumes {
			volumeList.WriteString(fmt.Sprintf("  - %s (%s)\n", volume.Name, volume.ID))
		}
	}

	if volumeList.Len() == 0 {
		return "(no volumes found)"
	}

	return volumeList.String()
}

// mapVolumeToState maps Redfish volume attributes to Terraform state.
func (d *StorageVolumeDataSource) mapVolumeToState(ctx context.Context, volume *redfish.Volume, data *storageVolumeDataSourceModel) {
	data.ID = types.StringValue(volume.ID)
	data.Name = types.StringValue(volume.Name)
	data.CapacityBytes = types.Int64Value(int64(volume.CapacityBytes))
	data.RAIDType = types.StringValue(string(volume.RAIDType))
	data.Status = types.StringValue(string(volume.Status.State))
	data.Encrypted = types.BoolValue(volume.Encrypted)

	// Controller ID - extract from volume ODataID or set empty
	// The controller ID is typically part of the volume's ODataID path
	data.ControllerID = types.StringValue("")
	if volume.ODataID != "" {
		// Extract controller from path like /redfish/v1/Systems/System.Embedded.1/Storage/RAID.Integrated.1-1/Volumes/Disk.Virtual.0:RAID.Integrated.1-1
		parts := strings.Split(volume.ODataID, "/")
		for i, part := range parts {
			if part == "Storage" && i+1 < len(parts) {
				data.ControllerID = types.StringValue(parts[i+1])
				break
			}
		}
	}

	// Creation time - not available in gofish Volume struct
	data.CreationTime = types.StringNull()

	// Physical disks - use Drives() method
	drives, err := volume.Drives()
	if err == nil && len(drives) > 0 {
		physicalDisks := make([]types.String, 0, len(drives))
		for _, drive := range drives {
			physicalDisks = append(physicalDisks, types.StringValue(drive.ODataID))
		}
		data.PhysicalDisks, _ = types.ListValueFrom(ctx, types.StringType, physicalDisks)
	} else {
		data.PhysicalDisks, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	// Optimum IO size
	if volume.OptimumIOSizeBytes > 0 {
		data.OptimumIOSize = types.Int64Value(int64(volume.OptimumIOSizeBytes))
	} else {
		data.OptimumIOSize = types.Int64Null()
	}

	// Block size
	if volume.BlockSizeBytes > 0 {
		data.BlockSize = types.Int64Value(int64(volume.BlockSizeBytes))
	} else {
		data.BlockSize = types.Int64Null()
	}

	tflog.Debug(ctx, "Mapped volume to state", map[string]any{
		"volume_id":   data.ID.ValueString(),
		"volume_name": data.Name.ValueString(),
	})
}
