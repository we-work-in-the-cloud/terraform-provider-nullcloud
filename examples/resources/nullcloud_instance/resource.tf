resource "nullcloud_vpc" "example" {
  name = "my-vpc"
}

resource "nullcloud_subnet" "example" {
  name       = "my-subnet"
  vpc_id     = nullcloud_vpc.example.id
  zone       = "us-east-1"
  cidr_block = "10.0.0.0/24"
}

resource "nullcloud_instance" "example" {
  name      = "my-vm"
  subnet_id = nullcloud_subnet.example.id
  profile   = "cx2-2x4"
  image     = "ubuntu-22-04"
}
