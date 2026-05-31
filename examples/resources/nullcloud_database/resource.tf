resource "nullcloud_database" "example" {
  name       = "my-db"
  engine     = "postgres"
  version    = "15"
  plan       = "medium"
  subnet_ids = [nullcloud_subnet.main.id]
}
