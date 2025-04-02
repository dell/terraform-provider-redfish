/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const importBuffer = `
import_buffer = base64decode(redfish_idrac_server_configuration_profile_export.config_1.file_content)
`

func getSP(shareType string, addOn map[string]string) string {
	parameters := make([]string, 0, len(addOn))
	for key, value := range addOn {
		parameters = append(parameters, fmt.Sprintf("%s = \"%s\"", key, value))
	}
	return fmt.Sprintf(`
	share_parameters = {
		filename   = "server_configuration_profile.xml"
		target     = ["EventFilters"]
		share_type = "%s"
		%s
	}
	`, shareType, strings.Join(parameters, "\n\t"))
}

func createSCPConfig(configType, configName, shareParameter, addOnConfig string) string {
	return fmt.Sprintf(`
		resource "redfish_idrac_server_configuration_profile_%s" "%s"  {
			redfish_server {
				user = "%s"
				password = "%s"
				endpoint = "%s"
				ssl_insecure = true
			  }
		  %s
		  %s
		}`,
		configType,
		configName,
		os.Getenv("TF_TESTING_USERNAME"),
		os.Getenv("TF_TESTING_PASSWORD"),
		os.Getenv("TF_TESTING_ENDPOINT"),
		shareParameter,
		addOnConfig,
	)
}

func dependsOnExport(configName string) string {
	return fmt.Sprintf(`
		depends_on = [
			redfish_idrac_server_configuration_profile_export.%s
		]`,
		configName,
	)
}

func TestAccRedfishSCP(t *testing.T) {
	nfs := map[string]string{
		"ip_address": os.Getenv("TF_NFS_IP_ADDRESS"),
		"share_name": os.Getenv("TF_NFS_SHARE_NAME"),
	}

	cifs := map[string]string{
		"ip_address": os.Getenv("TF_CIFS_IP_ADDRESS"),
		"share_name": os.Getenv("TF_CIFS_SHARE_NAME"),
		"username":   os.Getenv("TF_CIFS_USERNAME"),
		"password":   os.Getenv("TF_CIFS_PASSWORD"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				%s
				%s
				`,
					createSCPConfig("export", "config_1", getSP("LOCAL", nil), ""), createSCPConfig("import", "config_1", getSP("LOCAL", nil), importBuffer+"\n"+dependsOnExport("config_1"))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_idrac_server_configuration_profile_export.config_1", "share_parameters.share_type", "LOCAL"),
					resource.TestCheckResourceAttr("redfish_idrac_server_configuration_profile_import.config_1", "share_parameters.share_type", "LOCAL"),
				),
			},

			{
				Config: fmt.Sprintf(`
				%s
				%s
				`,
					createSCPConfig("export", "config_1", getSP("NFS", nfs), ""), createSCPConfig("import", "config_1", getSP("NFS", nfs), dependsOnExport("config_1"))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_idrac_server_configuration_profile_export.config_1", "share_parameters.share_type", "NFS"),
					resource.TestCheckResourceAttr("redfish_idrac_server_configuration_profile_import.config_1", "share_parameters.share_type", "NFS"),
				),
			},

			{
				Config: fmt.Sprintf(`
				%s
				%s
				`,
					createSCPConfig("export", "config_2", getSP("CIFS", cifs), ""), createSCPConfig("import", "config_2", getSP("CIFS", cifs), dependsOnExport("config_2"))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_idrac_server_configuration_profile_export.config_2", "share_parameters.share_type", "CIFS"),
					resource.TestCheckResourceAttr("redfish_idrac_server_configuration_profile_import.config_2", "share_parameters.share_type", "CIFS"),
				),
			},
		},
	})
}

func TestAccRedfishSCPInvalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{

			{
				Config:      createSCPConfig("export", "invalid_nfs", getSP("NFS", nil), ""),
				ExpectError: regexp.MustCompile("Export NFS Error"),
			},
			{
				Config:      createSCPConfig("import", "invalid_nfs", getSP("NFS", nil), ""),
				ExpectError: regexp.MustCompile("Import NFS Error"),
			},
			{
				Config:      createSCPConfig("export", "invalid_cifs", getSP("CIFS", nil), ""),
				ExpectError: regexp.MustCompile("Export CIFS Error"),
			},
			{
				Config:      createSCPConfig("import", "invalid_cifs", getSP("CIFS", nil), ""),
				ExpectError: regexp.MustCompile("Import CIFS Error"),
			},
			{
				Config:      createSCPConfig("export", "invalid_http", getSP("HTTP", nil), ""),
				ExpectError: regexp.MustCompile("Export HTTP/HTTPS IP Error"),
			},
			{
				Config:      createSCPConfig("import", "invalid_http", getSP("HTTP", nil), ""),
				ExpectError: regexp.MustCompile("Import HTTP/HTTPS IP Error"),
			},
			{
				Config:      createSCPConfig("export", "invalid_http_proxy", getSP("HTTP", map[string]string{"proxy_support": "true", "ip_address": "127.0.0.1"}), ""),
				ExpectError: regexp.MustCompile("Export Proxy Error"),
			},
			{
				Config:      createSCPConfig("import", "invalid_http_proxy", getSP("HTTP", map[string]string{"proxy_support": "true", "ip_address": "127.0.0.1"}), ""),
				ExpectError: regexp.MustCompile("Import Proxy Error"),
			},
		},
	})
}
