resource "nullcloud_cluster" "example" {
  name       = "my-cluster"
  version    = "1.30"
  node_count = 3
}
