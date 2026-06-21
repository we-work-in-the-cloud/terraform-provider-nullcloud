# Nullcloud List Resources Examples

This directory contains examples of using Terraform list resources to query and discover existing infrastructure in Nullcloud.

## Overview

List resources are a Terraform 1.9+ feature that enables querying existing infrastructure. Unlike data sources that require specific identifiers, list resources allow filtering and bulk discovery of resources.

List resources are invoked with the `terraform query` command using `.tfquery.hcl` files, not in regular `.tf` configuration files.

## Prerequisites

- Terraform 1.9 or later
- Nullcloud provider 0.5.0 or later
- Running Nullcloud backend API
- Valid Nullcloud provider credentials

## Running the Examples

1. Initialize Terraform (this will download required providers):
```bash
terraform init
```

2. Run a query to discover resources:
```bash
# Query resources and display results as JSON
terraform query -json

# Query and format as table (if supported)
terraform query
```

## Supported List Resources

### VPCs
Query VPCs with optional region filter:
```hcl
list "nullcloud_vpc" "example" {
  config {
    region = "us-east"  # Optional filter
  }
}
```

### Subnets
Query subnets with optional VPC and zone filters:
```hcl
list "nullcloud_subnet" "example" {
  config {
    vpc_id = "vpc-1"      # Optional filter
    zone   = "us-east-1"  # Optional filter
  }
}
```

### Instances
Query instances with optional subnet and status filters:
```hcl
list "nullcloud_instance" "example" {
  config {
    subnet_id = "subnet-1"  # Optional filter
    status    = "running"   # Optional filter: running, stopped, etc.
  }
}
```

### Load Balancers
Query load balancers with optional protocol filter:
```hcl
list "nullcloud_loadbalancer" "example" {
  config {
    protocol = "http"  # Optional filter: http, https, tcp, udp
  }
}
```

### Buckets
Query buckets with optional region filter:
```hcl
list "nullcloud_bucket" "example" {
  config {
    region = "us-west"  # Optional filter
  }
}
```

### Databases
Query databases with optional engine filter:
```hcl
list "nullcloud_database" "example" {
  config {
    engine = "postgres"  # Optional filter: postgres, mysql, etc.
  }
}
```

### Kubernetes Clusters
Query clusters with optional version filter:
```hcl
list "nullcloud_cluster" "example" {
  config {
    version = "1.24"  # Optional filter
  }
}
```

## Common Use Cases

### List all resources of a type
```hcl
list "nullcloud_vpc" "all" {
  config {}
}

output "vpc_ids" {
  value = list.nullcloud_vpc.all.results[*].id
}
```

### Filter resources by attribute
```hcl
list "nullcloud_vpc" "us_east" {
  config {
    region = "us-east"
  }
}

output "vpc_count" {
  value = length(list.nullcloud_vpc.us_east.results)
}
```

### Combine with local values for import
```hcl
locals {
  existing_vpcs = {
    for vpc in list.nullcloud_vpc.all.results :
    vpc.id => vpc
  }
}

output "vpc_map" {
  value = local.existing_vpcs
}
```

### Use query results to decide on imports
```hcl
locals {
  vpcs_to_import = [
    for vpc in list.nullcloud_vpc.all.results :
    vpc if vpc.region == "us-east"
  ]
}

output "import_commands" {
  value = [
    for vpc in local.vpcs_to_import :
    "terraform import nullcloud_vpc.${vpc.name} ${vpc.id}"
  ]
}
```

## Output Attributes

Each list resource result includes computed attributes. Common attributes:
- `id` - Unique resource identifier
- `name` - Resource name
- `status` - Current resource status
- `crn` - Cloud Resource Name
- `created_at` - Creation timestamp

Resource-specific attributes:
- **VPC**: `region`
- **Subnet**: `vpc_id`, `zone`, `cidr_block`
- **Instance**: `subnet_id`, `status`, `profile`, `image`, `primary_ip`
- **LoadBalancer**: `protocol`, `port`, `targets`
- **Bucket**: `region`
- **Database**: `engine`, `version`, `plan`, `endpoint`
- **Cluster**: `version`, `node_count`, `subnet_ids`

## More Information

For more information about Terraform list resources, see:
- [Terraform List Resources Documentation](https://developer.hashicorp.com/terraform/plugin/framework/list-resources)
- [Nullcloud Provider Documentation](https://registry.terraform.io/providers/we-work-in-the-cloud/nullcloud/latest)
