# NullCloud - Terraform provider

Terraform provider for [NullCloud](https://registry.terraform.io/providers/we-work-in-the-cloud/nullcloud).

> **Running the backend?** This provider talks to [backend-nullcloud](https://github.com/we-work-in-the-cloud/backend-nullcloud), a local fake-cloud API server. Start it before running Terraform.

## Resources

| Resource | Description |
|---|---|
| `nullcloud_vpc` | Virtual private cloud |
| `nullcloud_subnet` | Subnet within a VPC |
| `nullcloud_instance` | Virtual server instance |
| `nullcloud_loadbalancer` | Load balancer (tcp/http/https) |
| `nullcloud_bucket` | Object storage bucket |
| `nullcloud_database` | Managed database (postgres/mysql/mariadb) |
| `nullcloud_cluster` | Kubernetes cluster |

## Data Sources

| Data Source | Description |
|---|---|
| `nullcloud_vpc` | Fetch a VPC by ID |
| `nullcloud_subnet` | Fetch a subnet by ID |
| `nullcloud_instance` | Fetch a virtual server instance by ID |
| `nullcloud_loadbalancer` | Fetch a load balancer by ID |
| `nullcloud_bucket` | Fetch an object storage bucket by ID |
| `nullcloud_database` | Fetch a managed database by ID |
| `nullcloud_cluster` | Fetch a Kubernetes cluster by ID |

## Actions

| Action | Description |
|---|---|
| `nullcloud_instance_action` | Perform a `start`, `stop`, or `restart` on an instance |

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
