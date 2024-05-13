/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserAccountPassword struct to update password for user account with administrator privilige
type UserAccountPassword struct {
	ID          types.String `tfsdk:"id"`
	Endpoint    types.String `tfsdk:"endpoint"`
	SslInsecure types.Bool   `tfsdk:"ssl_insecure"`
	Username    types.String `tfsdk:"username"`
	OldPassword types.String `tfsdk:"old_password"`
	NewPassword types.String `tfsdk:"new_password"`
}
