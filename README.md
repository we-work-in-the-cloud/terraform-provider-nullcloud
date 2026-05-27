# terraform-provider-nullcloud

Terraform provider for [NullCloud](https://registry.terraform.io/providers/we-work-in-the-cloud/nullcloud).

## Resources

| Resource | Description |
|---|---|
| `nullcloud_vpc` | Virtual private cloud |
| `nullcloud_subnet` | Subnet within a VPC |
| `nullcloud_instance` | Virtual server instance |

## Usage

```hcl
terraform {
  required_providers {
    nullcloud = {
      source = "we-work-in-the-cloud/nullcloud"
    }
  }
}

provider "nullcloud" {
  url   = "http://localhost:8080"
  token = "mytoken"
}

resource "nullcloud_vpc" "main" {
  name = "my-vpc"
}

resource "nullcloud_subnet" "main" {
  name   = "my-subnet"
  vpc_id = nullcloud_vpc.main.id
}

resource "nullcloud_instance" "main" {
  name      = "my-vm"
  subnet_id = nullcloud_subnet.main.id
  profile   = "cx2-2x4"
  image     = "ubuntu-22-04"
}
```

## Provider configuration

| Argument | Required | Description |
|---|---|---|
| `url` | yes | NullCloud backend URL |
| `token` | yes | Authorization token |

## Development

```sh
make build    # cross-compile for all platforms into dist/
make install  # build and install for local OS/arch
make test     # run unit tests
```
