policytest {
  targets = ["../../policies/region/bucket-region.policy.hcl"]
}

resource "nullcloud_bucket" "allowed" {
  attrs = {
    region = "us-east"
  }
}

resource "nullcloud_bucket" "denied" {
  expect_failure = true
  attrs = {
    region = "us-west"
  }
}
