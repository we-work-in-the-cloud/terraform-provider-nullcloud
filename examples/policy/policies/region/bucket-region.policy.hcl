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
}

resource_policy "nullcloud_bucket" "bucket_region" {
  enforcement_level = "mandatory"
  enforce {
    condition     = attrs.region == "us-east"
    error_message = "Buckets must use region us-east."
  }
}
