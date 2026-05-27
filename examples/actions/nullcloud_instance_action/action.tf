action "nullcloud_instance_action" "stop" {
  config {
    instance_id = nullcloud_instance.example.id
    action      = "stop"
  }
}
