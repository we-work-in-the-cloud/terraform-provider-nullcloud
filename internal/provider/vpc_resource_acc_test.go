package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

// TestAccVPCResource_NameChange_NoDestroy verifies that changing the name does not destroy and recreate the resource.
func TestAccVPCResource_NameChange_NoDestroy(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	rName := RandomName("vpc")
	rNameUpdated := RandomName("vpc-renamed")
	var resourceID string

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
					// Capture the ID for comparison in next step
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_vpc.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						resourceID = rs.Primary.ID
						return nil
					},
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
					// Verify the ID has not changed (would change if resource was destroyed/recreated)
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_vpc.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						if rs.Primary.ID != resourceID {
							return fmt.Errorf("resource was destroyed/recreated: ID changed from %s to %s", resourceID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

