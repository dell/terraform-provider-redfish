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

const (

	// ServiceErrorMsg specifies error details occured while creating server connection
	ServiceErrorMsg = "service error"

	// RedfishAPIErrorMsg specifies error details occured while calling a redfish API
	RedfishAPIErrorMsg = "Error when contacting the redfish API"

	// RedfishFetchErrorMsg specifies error details occured while fetching details
	RedfishFetchErrorMsg = "Unable to fetch updated details"

	// RedfishVirtualMediaMountError specifies error when there are issues while mounting virtual media
	RedfishVirtualMediaMountError = "Couldn't mount Virtual Media"

	// RedfishJobErrorMsg specifies error details occured while tracking job details
	RedfishJobErrorMsg = "Error, job wasn't able to complete"

	// RedfishPasswordErrorMsg specifies if password validation fails in user resource
	RedfishPasswordErrorMsg = "Password validation failed"

	// Seventeen specifies the server generation for comparison
	Seventeen = 17
)
