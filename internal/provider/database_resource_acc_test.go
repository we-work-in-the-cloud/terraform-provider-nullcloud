package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccDatabaseResource_Create tests basic database creation.
func TestAccDatabaseResource_Create(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	dbName := RandomName("db")
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

	resource "nullcloud_database" "test" {
	  name       = "` + dbName + `"
	  engine     = "postgres"
	  version    = "15"
	  plan       = "small"
	  subnet_ids = [nullcloud_subnet.test.id]
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_database.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_database.test", "name", dbName),
					resource.TestCheckResourceAttr("nullcloud_database.test", "engine", "postgres"),
					resource.TestCheckResourceAttr("nullcloud_database.test", "version", "15"),
					resource.TestCheckResourceAttr("nullcloud_database.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccDatabaseResource_Update tests database update operations.
func TestAccDatabaseResource_Update(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	dbName := RandomName("db")
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

	resource "nullcloud_database" "test" {
	  name       = "` + dbName + `"
	  engine     = "postgres"
	  version    = "15"
	  plan       = "small"
	  subnet_ids = [nullcloud_subnet.test.id]
	}
	`,
				Check: resource.TestCheckResourceAttr("nullcloud_database.test", "version", "15"),
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

	resource "nullcloud_database" "test" {
	  name       = "` + dbName + `"
	  engine     = "postgres"
	  version    = "16"
	  plan       = "small"
	  subnet_ids = [nullcloud_subnet.test.id]
	}
	`,
				Check: resource.TestCheckResourceAttr("nullcloud_database.test", "version", "16"),
			},
		},
	})
}

// TestAccDatabaseResource_NameChange_NoDestroy verifies that changing the name does not destroy and recreate the resource.
func TestAccDatabaseResource_NameChange_NoDestroy(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")
	dbName := RandomName("db")
	dbNameUpdated := RandomName("db-renamed")
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

resource "nullcloud_database" "test" {
  name       = "` + dbName + `"
  engine     = "postgres"
  version    = "16"
  plan       = "small"
  subnet_ids = [nullcloud_subnet.test.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_database.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_database.test", "name", dbName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_database.test"]
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

resource "nullcloud_database" "test" {
  name       = "` + dbNameUpdated + `"
  engine     = "postgres"
  version    = "16"
  plan       = "small"
  subnet_ids = [nullcloud_subnet.test.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nullcloud_database.test", "name", dbNameUpdated),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_database.test"]
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
