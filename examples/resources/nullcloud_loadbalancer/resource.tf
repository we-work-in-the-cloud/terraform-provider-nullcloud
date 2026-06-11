resource "nullcloud_vpc" "example" {
  name = "my-vpc"
}

resource "nullcloud_subnet" "example" {
  name       = "my-subnet"
  vpc_id     = nullcloud_vpc.example.id
  zone       = "us-east-1"
  cidr_block = "10.0.0.0/24"
}

resource "nullcloud_cluster" "example" {
  name       = "my-cluster"
  version    = "1.30"
  node_count = 3
  subnet_ids = [nullcloud_subnet.example.id]
}

resource "nullcloud_loadbalancer" "example" {
  name     = "my-lb"
  protocol = "https"
  port     = 443

  targets = [
    {
      type = "cluster"
      id   = nullcloud_cluster.example.id
    }
  ]
}
