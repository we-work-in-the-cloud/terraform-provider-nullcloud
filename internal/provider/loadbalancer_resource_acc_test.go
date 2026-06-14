package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccLoadBalancerResource_Create tests basic load balancer creation.
func TestAccLoadBalancerResource_Create(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	lbName := RandomName("lb")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + vpcName + `"
  region = "us-east"
}

resource "nullcloud_loadbalancer" "test" {
  name     = "` + lbName + `"
  protocol = "tcp"
  port     = 80
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_loadbalancer.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "name", lbName),
					resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "port", "80"),
					resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "status", "active"),
				),
			},
		},
	})
}

// TestAccLoadBalancerResource_Update tests load balancer update operations.
func TestAccLoadBalancerResource_Update(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	lbName := RandomName("lb")
	lbNameUpdated := RandomName("lb-updated")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + vpcName + `"
  region = "us-east"
}

resource "nullcloud_loadbalancer" "test" {
  name     = "` + lbName + `"
  protocol = "tcp"
  port     = 80
}
`,
				Check: resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "name", lbName),
			},
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_vpc" "test" {
  name   = "` + vpcName + `"
  region = "us-east"
}

resource "nullcloud_loadbalancer" "test" {
  name     = "` + lbNameUpdated + `"
  protocol = "tcp"
  port     = 80
}
`,
				Check: resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "name", lbNameUpdated),
			},
		},
	})
}

// TestAccLoadBalancerResource_NameChange_NoDestroy verifies that changing the name does not destroy and recreate the resource.
func TestAccLoadBalancerResource_NameChange_NoDestroy(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	lbName := RandomName("lb")
	lbNameUpdated := RandomName("lb-renamed")
	var resourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_loadbalancer" "test" {
  name     = "` + lbName + `"
  protocol = "tcp"
  port     = 80
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_loadbalancer.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "name", lbName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_loadbalancer.test"]
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
resource "nullcloud_loadbalancer" "test" {
  name     = "` + lbNameUpdated + `"
  protocol = "tcp"
  port     = 80
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nullcloud_loadbalancer.test", "name", lbNameUpdated),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_loadbalancer.test"]
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
