resource "nullcloud_vpc" "example" {
  name = "my-vpc"
}

resource "nullcloud_subnet" "example" {
  name       = "my-subnet"
  vpc_id     = nullcloud_vpc.example.id
  zone       = "us-east-1"
  cidr_block = "10.0.0.0/24"
}

resource "nullcloud_database" "example" {
  name       = "my-db"
  engine     = "postgres"
  version    = "15"
  plan       = "medium"
  subnet_ids = [nullcloud_subnet.example.id]
}
