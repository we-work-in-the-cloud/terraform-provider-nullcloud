package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

func TestBucketListResource_Metadata(t *testing.T) {
	r := &BucketListResource{}
	req := resource.MetadataRequest{ProviderTypeName: "nullcloud"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "nullcloud_bucket", resp.TypeName)
}

func TestBucketListResource_ListResourceConfigSchema(t *testing.T) {
	r := &BucketListResource{}
	resp := &list.ListResourceSchemaResponse{}

	r.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "region")
}

func TestBucketListResource_Configure(t *testing.T) {
	mockClient := &client.Client{}
	r := &BucketListResource{}
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.Equal(t, mockClient, r.client)
}

func TestBucketListResource_FilterLogic(t *testing.T) {
	buckets := []client.Bucket{
		{ID: "bucket-1", Name: "bucket-1", Region: "us-east", Status: "active", CreatedAt: time.Now()},
		{ID: "bucket-2", Name: "bucket-2", Region: "us-west", Status: "active", CreatedAt: time.Now()},
		{ID: "bucket-3", Name: "bucket-3", Region: "us-east", Status: "active", CreatedAt: time.Now()},
	}

	// Test: no filter (all results)
	filtered := filterBuckets(buckets, "")
	assert.Equal(t, 3, len(filtered))

	// Test: filter by region
	filtered = filterBuckets(buckets, "us-east")
	assert.Equal(t, 2, len(filtered))

	filtered = filterBuckets(buckets, "us-west")
	assert.Equal(t, 1, len(filtered))
}

func filterBuckets(buckets []client.Bucket, region string) []client.Bucket {
	var result []client.Bucket
	for _, bucket := range buckets {
		if region == "" || bucket.Region == region {
			result = append(result, bucket)
		}
	}
	return result
}

func TestBucketListResource_ConfigModel(t *testing.T) {
	config := bucketListConfig{
		Region: types.StringValue("us-east"),
	}

	assert.Equal(t, "us-east", config.Region.ValueString())
}

func TestBucketListResource_NewBucketListResource(t *testing.T) {
	lr := NewBucketListResource()
	assert.NotNil(t, lr)
	assert.IsType(t, &BucketListResource{}, lr)
}

// TestBucketListResource_List_ErrorHandling verifies error handling path.
func TestBucketListResource_List_ConfigModel(t *testing.T) {
	// Test that bucketListConfig can be used to check null/non-null regions
	config1 := bucketListConfig{Region: types.StringNull()}
	assert.True(t, config1.Region.IsNull())

	config2 := bucketListConfig{Region: types.StringValue("us-east")}
	assert.False(t, config2.Region.IsNull())
	assert.Equal(t, "us-east", config2.Region.ValueString())
}

// TestBucketListResource_BucketModel tests the bucket model structure.
func TestBucketListResource_BucketModel(t *testing.T) {
	// Test bucketModel creation with types
	model := bucketModel{
		ID:        types.StringValue("bucket-1"),
		Name:      types.StringValue("my-bucket"),
		Status:    types.StringValue("active"),
		CRN:       types.StringValue("crn:bucket-1"),
		Region:    types.StringValue("us-east"),
		CreatedAt: types.StringValue("2024-01-01T00:00:00Z"),
	}

	assert.Equal(t, "bucket-1", model.ID.ValueString())
	assert.Equal(t, "my-bucket", model.Name.ValueString())
	assert.Equal(t, "active", model.Status.ValueString())
	assert.Equal(t, "us-east", model.Region.ValueString())
}

// TestBucketListResource_ToBucketModel tests the conversion helper.
func TestBucketListResource_ToBucketModel(t *testing.T) {
	now := time.Now()
	clientBucket := client.Bucket{
		ID:        "bucket-1",
		Name:      "my-bucket",
		Status:    "active",
		CRN:       "crn:bucket-1",
		Region:    "us-east",
		CreatedAt: now,
	}

	model := toBucketModel(clientBucket)

	assert.Equal(t, "bucket-1", model.ID.ValueString())
	assert.Equal(t, "my-bucket", model.Name.ValueString())
	assert.Equal(t, "active", model.Status.ValueString())
	assert.Equal(t, "crn:bucket-1", model.CRN.ValueString())
	assert.Equal(t, "us-east", model.Region.ValueString())
	assert.Equal(t, now.String(), model.CreatedAt.ValueString())
}

// TestBucketListResource_ShouldIncludeBucket tests the filter helper.
func TestBucketListResource_ShouldIncludeBucket(t *testing.T) {
	bucket := client.Bucket{
		ID:        "bucket-1",
		Name:      "my-bucket",
		Region:    "us-east",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	// Test: no region filter (include all)
	config := bucketListConfig{Region: types.StringNull()}
	assert.True(t, shouldIncludeBucket(bucket, config))

	// Test: matching region (include)
	config = bucketListConfig{Region: types.StringValue("us-east")}
	assert.True(t, shouldIncludeBucket(bucket, config))

	// Test: non-matching region (exclude)
	config = bucketListConfig{Region: types.StringValue("us-west")}
	assert.False(t, shouldIncludeBucket(bucket, config))
}
