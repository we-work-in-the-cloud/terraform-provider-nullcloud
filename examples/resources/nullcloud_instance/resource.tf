resource "nullcloud_instance" "example" {
  name      = "my-vm"
  subnet_id = nullcloud_subnet.example.id
  profile   = "cx2-2x4"
  image     = "ubuntu-22-04"
}
