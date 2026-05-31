resource "nullcloud_cluster" "example" {
  name       = "my-cluster"
  version    = "1.30"
  node_count = 3
  subnet_ids = [nullcloud_subnet.main.id]
}
