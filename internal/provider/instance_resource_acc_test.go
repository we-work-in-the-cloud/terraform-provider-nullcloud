package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccInstanceResource_Create tests basic instance creation.
func TestAccInstanceResource_Create(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")
	instanceName := RandomName("instance")

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

resource "nullcloud_instance" "test" {
  name      = "` + instanceName + `"
  subnet_id = nullcloud_subnet.test.id
  profile   = "bx2-2x8"
  image     = "ibm-ubuntu-22-04"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_instance.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_instance.test", "name", instanceName),
					resource.TestCheckResourceAttr("nullcloud_instance.test", "profile", "bx2-2x8"),
					resource.TestCheckResourceAttr("nullcloud_instance.test", "image", "ibm-ubuntu-22-04"),
					resource.TestCheckResourceAttr("nullcloud_instance.test", "status", "running"),
					resource.TestCheckResourceAttrSet("nullcloud_instance.test", "primary_ip"),
				),
			},
		},
	})
}

// TestAccInstanceResource_Update tests instance profile change (requires replacement).
func TestAccInstanceResource_Update(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")
	instanceName := RandomName("instance")

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

resource "nullcloud_instance" "test" {
  name      = "` + instanceName + `"
  subnet_id = nullcloud_subnet.test.id
  profile   = "bx2-2x8"
  image     = "ibm-ubuntu-22-04"
}
`,
				Check: resource.TestCheckResourceAttr("nullcloud_instance.test", "profile", "bx2-2x8"),
			},
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

resource "nullcloud_instance" "test" {
  name      = "` + instanceName + `"
  subnet_id = nullcloud_subnet.test.id
  profile   = "bx2-4x16"
  image     = "ibm-ubuntu-22-04"
}
`,
				Check: resource.TestCheckResourceAttr("nullcloud_instance.test", "profile", "bx2-4x16"),
			},
		},
	})
}

// TestAccInstanceResource_NameChange_NoDestroy verifies that changing the name does not destroy and recreate the resource.
func TestAccInstanceResource_NameChange_NoDestroy(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")
	instanceName := RandomName("instance")
	instanceNameUpdated := RandomName("instance-renamed")
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

resource "nullcloud_instance" "test" {
  name    = "` + instanceName + `"
  subnet_id = nullcloud_subnet.test.id
  profile = "cx2-2x4"
  image   = "ubuntu-22.04"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_instance.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_instance.test", "name", instanceName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_instance.test"]
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
  name       = "` + subnetName + `"
  vpc_id     = nullcloud_vpc.test.id
  zone       = "us-east-1"
  cidr_block = "10.0.0.0/24"
}

resource "nullcloud_instance" "test" {
  name    = "` + instanceNameUpdated + `"
  subnet_id = nullcloud_subnet.test.id
  profile = "cx2-2x4"
  image   = "ubuntu-22.04"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nullcloud_instance.test", "name", instanceNameUpdated),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_instance.test"]
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
