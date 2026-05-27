terraform {
  required_version = ">= 1.14"
  required_providers {
    nullcloud = {
      source = "registry.terraform.io/we-work-in-the-cloud/nullcloud"
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
