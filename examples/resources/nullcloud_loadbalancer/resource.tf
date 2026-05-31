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
