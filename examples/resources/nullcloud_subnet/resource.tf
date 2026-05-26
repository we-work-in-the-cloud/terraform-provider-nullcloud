resource "nullcloud_subnet" "example" {
  name   = "my-subnet"
  vpc_id = nullcloud_vpc.example.id
}
