/*
Copyright (c) 2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package helper

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	timeBetweenAttemptsCatalogUpdate = 20
	timeoutForCatalogUpdate          = 720
	defaultPort                      = 80
	httpString                       = "HTTP"
)

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
		if !updateObject.IsNull() {
			updateObjects = append(updateObjects, updateObject)
		}
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
			return nil, fmt.Errorf("proxy is only supported when ShareType is HTTP, HTTPS or FTP")
		}
		if plan.ProxyServer.ValueString() == "" {
			return nil, fmt.Errorf("proxy_server is required when ProxySupport is Enabled")
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
