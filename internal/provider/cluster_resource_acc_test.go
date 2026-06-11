package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccClusterResource_Create tests basic Kubernetes cluster creation.
func TestAccClusterResource_Create(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")
	clusterName := RandomName("cluster")

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

	resource "nullcloud_cluster" "test" {
	  name       = "` + clusterName + `"
	  subnet_ids = [nullcloud_subnet.test.id]
	  version    = "1.28"
	  node_count = 3
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_cluster.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("nullcloud_cluster.test", "version", "1.28"),
					resource.TestCheckResourceAttr("nullcloud_cluster.test", "node_count", "3"),
					resource.TestCheckResourceAttr("nullcloud_cluster.test", "status", "running"),
				),
			},
		},
	})
}

// TestAccClusterResource_Update tests cluster update operations.
func TestAccClusterResource_Update(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")
	subnetName := RandomName("subnet")
	clusterName := RandomName("cluster")

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

	resource "nullcloud_cluster" "test" {
	  name       = "` + clusterName + `"
	  subnet_ids = [nullcloud_subnet.test.id]
	  version    = "1.28"
	  node_count = 3
	}
	`,
				Check: resource.TestCheckResourceAttr("nullcloud_cluster.test", "node_count", "3"),
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

	resource "nullcloud_cluster" "test" {
	  name       = "` + clusterName + `"
	  subnet_ids = [nullcloud_subnet.test.id]
	  version    = "1.28"
	  node_count = 5
	}
	`,
				Check: resource.TestCheckResourceAttr("nullcloud_cluster.test", "node_count", "5"),
			},
		},
	})
}
