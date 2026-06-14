package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccSubnetResource_Create tests basic subnet creation.
func TestAccSubnetResource_Create(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + vpcName + `"
  region = "us-east"
}

resource "nullcloud_subnet" "test" {
  name      = "` + subnetName + `"
  vpc_id    = nullcloud_vpc.test.id
  zone      = "us-east-1"
  cidr_block = "10.0.1.0/24"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_subnet.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_subnet.test", "name", subnetName),
					resource.TestCheckResourceAttr("nullcloud_subnet.test", "zone", "us-east-1"),
					resource.TestCheckResourceAttr("nullcloud_subnet.test", "cidr_block", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("nullcloud_subnet.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccSubnetResource_Delete tests subnet deletion.
func TestAccSubnetResource_Delete(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + vpcName + `"
  region = "us-east"
}

resource "nullcloud_subnet" "test" {
  name       = "` + subnetName + `"
  vpc_id     = nullcloud_vpc.test.id
  zone       = "us-east-1"
  cidr_block = "10.0.1.0/24"
}
`,
				Check: resource.TestCheckResourceAttrSet("nullcloud_subnet.test", "id"),
			},
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + vpcName + `"
  region = "us-east"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

// TestAccSubnetResource_NameChange_NoDestroy verifies that changing the name does not destroy and recreate the resource.
func TestAccSubnetResource_NameChange_NoDestroy(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")
	subnetNameUpdated := RandomName("subnet-renamed")
	var resourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + vpcName + `"
  region = "us-east"
}

resource "nullcloud_subnet" "test" {
  name       = "` + subnetName + `"
  vpc_id     = nullcloud_vpc.test.id
  zone       = "us-east-1"
  cidr_block = "10.0.0.0/24"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_subnet.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_subnet.test", "name", subnetName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_subnet.test"]
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
  name   = "` + vpcName + `"
  region = "us-east"
}

resource "nullcloud_subnet" "test" {
  name       = "` + subnetNameUpdated + `"
  vpc_id     = nullcloud_vpc.test.id
  zone       = "us-east-1"
  cidr_block = "10.0.0.0/24"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nullcloud_subnet.test", "name", subnetNameUpdated),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_subnet.test"]
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
