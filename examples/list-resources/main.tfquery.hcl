# List Resources Example
#
# List resources enable discovery and filtering of existing infrastructure using
# the `terraform query` command. Use these examples to learn how to query resources
# in your Terraform configuration.

# Query all VPCs
list "nullcloud_vpc" "all_vpcs" {
  provider = nullcloud
  config {}
}

# Query VPCs in a specific region
list "nullcloud_vpc" "us_east_vpcs" {
  provider = nullcloud
  config {
    region = "us-east"
  }
}

# Query all subnets
list "nullcloud_subnet" "all_subnets" {
  provider = nullcloud
  config {}
}

# Query subnets in a specific VPC and zone
list "nullcloud_subnet" "vpc1_subnets" {
  provider = nullcloud
  config {
    vpc_id = "vpc-1"
    zone   = "us-east-1"
  }
}

# Query all instances
list "nullcloud_instance" "all_instances" {
  provider = nullcloud
  config {}
}

# Query running instances in a specific subnet
list "nullcloud_instance" "running_in_subnet" {
  provider = nullcloud
  config {
    subnet_id = "subnet-1"
    status    = "running"
  }
}

# Query all load balancers
list "nullcloud_loadbalancer" "all_lbs" {
  provider = nullcloud
  config {}
}

# Query HTTP load balancers
list "nullcloud_loadbalancer" "http_lbs" {
  provider = nullcloud
  config {
    protocol = "http"
  }
}

# Query all buckets
list "nullcloud_bucket" "all_buckets" {
  provider = nullcloud
  config {}
}

# Query buckets in a specific region
list "nullcloud_bucket" "us_west_buckets" {
  provider = nullcloud
  config {
    region = "us-west"
  }
}

# Query all databases
list "nullcloud_database" "all_databases" {
  provider = nullcloud
  config {}
}

# Query PostgreSQL databases
list "nullcloud_database" "postgres_databases" {
  provider = nullcloud
  config {
    engine = "postgres"
  }
}

# Query all Kubernetes clusters
list "nullcloud_cluster" "all_clusters" {
  provider = nullcloud
  config {}
}

# Query clusters running Kubernetes 1.24
list "nullcloud_cluster" "k8s_1_24_clusters" {
  provider = nullcloud
  config {
    version = "1.24"
  }
}
