package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
