policy {
  required_providers {
    nullcloud = {
      source  = "we-work-in-the-cloud/nullcloud"
      version = "0.6.0"
    }
  }
  terraform_config {
    required_version = ">= 1.16.0"
  }
  plugins {
    nullcloud = {
      source = "../../plugins/bin/nullcloud-policy"
    }
  }
}

resource_policy "nullcloud_bucket" "bucket_crn" {
  enforcement_level = "mandatory"
  enforce {
    condition     = plugin::nullcloud::valid_crn(attrs.crn, "bucket")
    error_message = "Buckets must have a valid NullCloud bucket CRN."
  }
}
