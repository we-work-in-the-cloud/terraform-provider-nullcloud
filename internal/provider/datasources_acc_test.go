package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccVPCDataSource tests reading VPC data source.
func TestAccVPCDataSource(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	vpcName := RandomName("vpc")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
	resource "nullcloud_vpc" "test" {
	  name   = "` + vpcName + `"
	  region = "us-east"
	}

	data "nullcloud_vpc" "test" {
	  id = nullcloud_vpc.test.id
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nullcloud_vpc.test", "name", vpcName),
					resource.TestCheckResourceAttr("data.nullcloud_vpc.test", "region", "us-east"),
					resource.TestCheckResourceAttr("data.nullcloud_vpc.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccSubnetDataSource tests reading subnet data source.
func TestAccSubnetDataSource(t *testing.T) {
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

	data "nullcloud_subnet" "test" {
	  id = nullcloud_subnet.test.id
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nullcloud_subnet.test", "name", subnetName),
					resource.TestCheckResourceAttr("data.nullcloud_subnet.test", "zone", "us-east-1"),
					resource.TestCheckResourceAttr("data.nullcloud_subnet.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccInstanceDataSource tests reading instance data source.
func TestAccInstanceDataSource(t *testing.T) {
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

	data "nullcloud_instance" "test" {
	  id = nullcloud_instance.test.id
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nullcloud_instance.test", "name", instanceName),
					resource.TestCheckResourceAttr("data.nullcloud_instance.test", "status", "running"),
					resource.TestCheckResourceAttrSet("data.nullcloud_instance.test", "primary_ip"),
				),
			},
		},
	})
}

// TestAccBucketDataSource tests reading bucket data source.
func TestAccBucketDataSource(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	bucketName := RandomName("bucket")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
	resource "nullcloud_bucket" "test" {
	  name   = "` + bucketName + `"
	  region = "us-east"
	}

	data "nullcloud_bucket" "test" {
	  id = nullcloud_bucket.test.id
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nullcloud_bucket.test", "name", bucketName),
					resource.TestCheckResourceAttr("data.nullcloud_bucket.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccDatabaseDataSource tests reading database data source.
func TestAccDatabaseDataSource(t *testing.T) {
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

	data "nullcloud_database" "test" {
	  id = nullcloud_database.test.id
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nullcloud_database.test", "name", dbName),
					resource.TestCheckResourceAttr("data.nullcloud_database.test", "engine", "postgres"),
					resource.TestCheckResourceAttr("data.nullcloud_database.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccLoadBalancerDataSource tests reading load balancer data source.
func TestAccLoadBalancerDataSource(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	lbName := RandomName("lb")

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

	data "nullcloud_loadbalancer" "test" {
	  id = nullcloud_loadbalancer.test.id
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nullcloud_loadbalancer.test", "name", lbName),
					resource.TestCheckResourceAttr("data.nullcloud_loadbalancer.test", "status", "active"),
				),
			},
		},
	})
}

// TestAccClusterDataSource tests reading cluster data source.
func TestAccClusterDataSource(t *testing.T) {
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

	data "nullcloud_cluster" "test" {
	  id = nullcloud_cluster.test.id
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nullcloud_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("data.nullcloud_cluster.test", "status", "running"),
				),
			},
		},
	})
}

// TestAccRegionsDataSource tests reading regions data source.
func TestAccRegionsDataSource(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
	data "nullcloud_regions" "all" {
	}
	`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.nullcloud_regions.all", "regions.#"),
				),
			},
		},
	})
}
