resource "nullcloud_database" "example" {
  name    = "my-db"
  engine  = "postgres"
  version = "15"
  plan    = "medium"
}
