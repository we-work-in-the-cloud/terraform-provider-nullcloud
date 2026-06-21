terraform {
  required_version = ">= 1.9"
  required_providers {
    nullcloud = {
      source  = "we-work-in-the-cloud/nullcloud"
      version = ">= 0.5.0"
    }
  }
}

provider "nullcloud" {
  url   = "http://localhost:8080"
  token = "test-token"
}
