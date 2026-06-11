package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccVPCResource_Create tests basic VPC creation with the live backend.
func TestAccVPCResource_Create(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	rName := RandomName("vpc")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + rName + `"
  region = "us-east"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_vpc.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_vpc.test", "name", rName),
					resource.TestCheckResourceAttr("nullcloud_vpc.test", "region", "us-east"),
					resource.TestCheckResourceAttr("nullcloud_vpc.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccVPCResource_Update tests VPC update operations.
func TestAccVPCResource_Update(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	rName := RandomName("vpc")
	rNameUpdated := RandomName("vpc-updated")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + rName + `"
  region = "us-east"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nullcloud_vpc.test", "name", rName),
				),
			},
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + rNameUpdated + `"
  region = "us-east"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nullcloud_vpc.test", "name", rNameUpdated),
				),
			},
		},
	})
}

