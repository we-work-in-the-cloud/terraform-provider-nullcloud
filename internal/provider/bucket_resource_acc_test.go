package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
