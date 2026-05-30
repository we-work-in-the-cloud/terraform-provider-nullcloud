resource "nullcloud_loadbalancer" "example" {
  name     = "my-lb"
  protocol = "https"
  port     = 443
}
