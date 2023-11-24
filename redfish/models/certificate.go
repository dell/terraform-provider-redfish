package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SSLCertificate struct for payload construct for create certificate api
type SSLCertificate struct {
	CertificateType    string `json:"CertificateType"`
	Passphrase         string `json:"Passphrase"`
	SSLCertificateFile string `json:"SSLCertificateFile"`
}

// RedfishSSLCertificate for terraform schema of certificate resource
type RedfishSSLCertificate struct {
	ID                 types.String    `tfsdk:"id"`
	RedfishServer      []RedfishServer `tfsdk:"redfish_server"`
	CertificateType    types.String    `tfsdk:"certificate_type"`
	Passphrase         types.String    `tfsdk:"passphrase"`
	SSLCertificateFile types.String    `tfsdk:"ssl_certificate_content"`
}