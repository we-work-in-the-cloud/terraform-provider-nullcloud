terraform {
  required_version = ">= 1.14"
  required_providers {
    nullcloud = {
      source = "we-work-in-the-cloud/nullcloud"
    }
  }
}

provider "nullcloud" {
  url   = "http://localhost:8080"
  token = "mytoken"
}

resource "nullcloud_vpc" "main" {
  name = "my-vpc"
}

resource "nullcloud_subnet" "main" {
  name   = "my-subnet"
  vpc_id = nullcloud_vpc.main.id
}

resource "nullcloud_instance" "main" {
  name      = "my-vm"
  subnet_id = nullcloud_subnet.main.id
  profile   = "cx2-2x4"
  image     = "ubuntu-22-04"
}

resource "nullcloud_loadbalancer" "main" {
  name     = "my-lb"
  protocol = "https"
  port     = 443

  targets = [
    {
      type = "vsi"
      id   = nullcloud_instance.main.id
    }
  ]
}

resource "nullcloud_bucket" "main" {
  name   = "my-bucket"
  region = "us-east-1"
}

resource "nullcloud_database" "main" {
  name       = "my-db"
  engine     = "postgres"
  version    = "15"
  plan       = "medium"
  subnet_ids = [nullcloud_subnet.main.id]
}

resource "nullcloud_cluster" "main" {
  name       = "my-cluster"
  version    = "1.30"
  node_count = 3
  subnet_ids = [nullcloud_subnet.main.id]
}

action "nullcloud_instance_action" "stop" {
  config {
    instance_id = nullcloud_instance.main.id
    action      = "stop"
  }
}

data "nullcloud_vpc" "main" {
  id = nullcloud_vpc.main.id
}

data "nullcloud_subnet" "main" {
  id = nullcloud_subnet.main.id
}

data "nullcloud_instance" "main" {
  id = nullcloud_instance.main.id
}

data "nullcloud_loadbalancer" "main" {
  id = nullcloud_loadbalancer.main.id
}

data "nullcloud_bucket" "main" {
  id = nullcloud_bucket.main.id
}

data "nullcloud_database" "main" {
  id = nullcloud_database.main.id
}

data "nullcloud_cluster" "main" {
  id = nullcloud_cluster.main.id
}

output "vpc_id" {
  value = nullcloud_vpc.main.id
}

output "vpc_crn" {
  value = data.nullcloud_vpc.main.crn
}

output "subnet_cidr" {
  value = nullcloud_subnet.main.cidr_block
}

output "instance_ip" {
  value = nullcloud_instance.main.primary_ip
}

output "instance_status" {
  value = nullcloud_instance.main.status
}

output "lb_crn" {
  value = nullcloud_loadbalancer.main.crn
}

output "bucket_region" {
  value = nullcloud_bucket.main.region
}

output "database_engine" {
  value = nullcloud_database.main.engine
}

output "cluster_version" {
  value = nullcloud_cluster.main.version
}
