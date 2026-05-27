---
page_title: "nullcloud Provider"
description: |-
  The NullCloud provider manages VPCs, subnets, and virtual server instances on NullCloud.
---

# nullcloud Provider

The NullCloud provider manages VPCs, subnets, and virtual server instances on NullCloud.

## Example Usage

```terraform
provider "nullcloud" {
  url   = "http://localhost:8080"
  token = "mytoken"
}
```

## Schema

### Required

- `token` (String, Sensitive) Authorization token
- `url` (String) NullCloud backend URL, e.g. http://localhost:8080

## Resources

- [nullcloud_vpc](resources/vpc.md) — Manages a VPC.
- [nullcloud_subnet](resources/subnet.md) — Manages a subnet within a VPC.
- [nullcloud_instance](resources/instance.md) — Manages a virtual server instance.

## Actions

- [nullcloud_instance_action](actions/instance_action.md) — Performs a start, stop, or restart action on an instance.

## Data Sources

- [nullcloud_vpc](data-sources/vpc.md) — Fetches a VPC by ID.
- [nullcloud_subnet](data-sources/subnet.md) — Fetches a subnet by ID.
- [nullcloud_instance](data-sources/instance.md) — Fetches a virtual server instance by ID.
