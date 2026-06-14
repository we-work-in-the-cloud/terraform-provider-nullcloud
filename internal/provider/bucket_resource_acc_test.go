package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccBucketResource_Create tests basic bucket creation.
func TestAccBucketResource_Create(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	bucketName := RandomName("bucket")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_bucket" "test" {
  name   = "` + bucketName + `"
  region = "us-east"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_bucket.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_bucket.test", "name", bucketName),
					resource.TestCheckResourceAttr("nullcloud_bucket.test", "region", "us-east"),
					resource.TestCheckResourceAttr("nullcloud_bucket.test", "status", "available"),
				),
			},
		},
	})
}

// TestAccBucketResource_Delete tests bucket deletion.
func TestAccBucketResource_Delete(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	bucketName := RandomName("bucket")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_bucket" "test" {
  name   = "` + bucketName + `"
  region = "us-east"
}
`,
				Check: resource.TestCheckResourceAttrSet("nullcloud_bucket.test", "id"),
			},
			{
				Config: bm.GetProviderConfig(),
			},
		},
	})
}

// TestAccBucketResource_NameChange_NoDestroy verifies that changing the name does not destroy and recreate the resource.
func TestAccBucketResource_NameChange_NoDestroy(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	bucketName := RandomName("bucket")
	bucketNameUpdated := RandomName("bucket-renamed")
	var resourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_bucket" "test" {
  name   = "` + bucketName + `"
  region = "us-east"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("nullcloud_bucket.test", "id"),
					resource.TestCheckResourceAttr("nullcloud_bucket.test", "name", bucketName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_bucket.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						resourceID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: bm.GetProviderConfig() + `
resource "nullcloud_bucket" "test" {
  name   = "` + bucketNameUpdated + `"
  region = "us-east"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nullcloud_bucket.test", "name", bucketNameUpdated),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["nullcloud_bucket.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						if rs.Primary.ID != resourceID {
							return fmt.Errorf("resource was destroyed/recreated: ID changed from %s to %s", resourceID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}
