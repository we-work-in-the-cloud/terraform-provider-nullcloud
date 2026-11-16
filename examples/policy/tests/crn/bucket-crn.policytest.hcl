policytest {
  targets = ["../../policies/crn/bucket-crn.policy.hcl"]
}

resource "nullcloud_bucket" "valid" {
  attrs = {
    crn = "crn:nullcloud:bucket:bucket-1"
  }
}

resource "nullcloud_bucket" "mismatched" {
  expect_failure = true
  attrs = {
    crn = "crn:nullcloud:vpc:vpc-1"
  }
}

resource "nullcloud_bucket" "malformed" {
  expect_failure = true
  attrs = {
    crn = "crn:nullcloud:bucket:"
  }
}
