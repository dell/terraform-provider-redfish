cd .././..
make install
cd examples/resources/redfish_boot_order/
rm -rf .terraform*  /tmp/log terraform.tfstate
terraform init
terraform apply
