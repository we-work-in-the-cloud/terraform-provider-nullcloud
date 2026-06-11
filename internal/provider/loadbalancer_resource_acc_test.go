package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
