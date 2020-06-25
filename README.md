# Terraform provider for Redfish
The Terraform provider for Redfish is a plugin for Terraform that allows for full lifecycle management of x86 servers using Redfish REST APIs. For more details on Redfish, please refer to DMTF Redfish specification [here][redfish-website].

For general information about Terraform, visit the [official website][tf-website] and the [GitHub project page][tf-github].

[redfish-website]: https://www.dmtf.org/standards/redfish
[tf-website]: https://terraform.io
[tf-github]: https://github.com/hashicorp/terraform

## Requirements
-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.13.x (to build the provider plugin)

## Installation

*Note*: This project uses [Go modules](https://blog.golang.org/using-go-modules) making it safe to work with it outside of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH).  The instructions that follow assume a directory in your home directory outside of the standard GOPATH (i.e `$HOME/development/terraform-providers/`).

Clone repository to: `$HOME/development/terraform-providers/`
```sh
$ mkdir -p $HOME/development/terraform-providers/; cd $HOME/development/terraform-providers/
$ git clone https://github.com/dell/terraform-provider-redfish.git
...
```

Enter the provider directory and run `go build`. This will build the provider and put the provider in the `$GOPATH/bin` directory.
```sh
$ go build
...
$ $GOPATH/bin/terraform-provider-redfish
...
```

## Documentation
The documentation for the provider can found here - Coming soon

## Roadmap
Our roadmap for Terraform provider for Redfish resources can be found [here](ROADMAP.md)

## Support
The code is provided AS-IS and not supported by Dell EMC.

## Contributing
The Terrafrom Redfish provider is open-source and community supported. We appreciate your help!
To contribute, please read the [contribution guidelines](docs/CONTRIBUTING.md). You may also [report an issue](https://github.com/dell/terraform-provider-redfish/issues/new/choose). Once you've filed an issue, it will follow the [issue lifecycle](docs/ISSUES.md).
