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

// TestVPCListResource_Metadata verifies the list resource declares the correct type name.
func TestVPCListResource_Metadata(t *testing.T) {
	r := &VPCListResource{}
	req := resource.MetadataRequest{
		ProviderTypeName: "nullcloud",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "nullcloud_vpc", resp.TypeName)
}

// TestVPCListResource_ListResourceConfigSchema verifies the list resource schema.
func TestVPCListResource_ListResourceConfigSchema(t *testing.T) {
	r := &VPCListResource{}
	resp := &list.ListResourceSchemaResponse{}

	r.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema)
	assert.NotNil(t, resp.Schema.Attributes)
	assert.Contains(t, resp.Schema.Attributes, "region")
}

// TestVPCListResource_Configure sets the client.
func TestVPCListResource_Configure(t *testing.T) {
	mockClient := &client.Client{}
	r := &VPCListResource{}
	req := resource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.Equal(t, mockClient, r.client)
}

// TestVPCListResource_Configure_NoData handles nil provider data.
func TestVPCListResource_Configure_NoData(t *testing.T) {
	r := &VPCListResource{}
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	// Should not panic or error
	r.Configure(context.Background(), req, resp)
	assert.Nil(t, r.client)
}

// TestVPCListResource_FilterLogic tests the filtering logic in isolation.
func TestVPCListResource_FilterLogic(t *testing.T) {
	vpcs := []client.VPC{
		{ID: "vpc-1", Name: "vpc-1", Region: "us-east", Status: "available", CRN: "crn:vpc-1", CreatedAt: time.Now()},
		{ID: "vpc-2", Name: "vpc-2", Region: "us-west", Status: "available", CRN: "crn:vpc-2", CreatedAt: time.Now()},
		{ID: "vpc-3", Name: "vpc-3", Region: "us-east", Status: "available", CRN: "crn:vpc-3", CreatedAt: time.Now()},
	}

	// Test: no filter (all results)
	filtered := filterVPCs(vpcs, "")
	assert.Equal(t, 3, len(filtered))

	// Test: filter by us-east
	filtered = filterVPCs(vpcs, "us-east")
	assert.Equal(t, 2, len(filtered))
	assert.Equal(t, "vpc-1", filtered[0].ID)
	assert.Equal(t, "vpc-3", filtered[1].ID)

	// Test: filter by us-west
	filtered = filterVPCs(vpcs, "us-west")
	assert.Equal(t, 1, len(filtered))
	assert.Equal(t, "vpc-2", filtered[0].ID)

	// Test: filter by non-existent region
	filtered = filterVPCs(vpcs, "eu-west")
	assert.Equal(t, 0, len(filtered))
}

// filterVPCs applies region filter to VPCs (mimics the List() logic).
func filterVPCs(vpcs []client.VPC, region string) []client.VPC {
	var result []client.VPC
	for _, vpc := range vpcs {
		if region == "" || vpc.Region == region {
			result = append(result, vpc)
		}
	}
	return result
}

// TestVPCListResource_ConfigModel tests vpcListConfig parsing.
func TestVPCListResource_ConfigModel(t *testing.T) {
	config := vpcListConfig{
		Region: types.StringValue("us-east"),
	}

	assert.Equal(t, "us-east", config.Region.ValueString())
	assert.False(t, config.Region.IsNull())

	// Test null region
	nullConfig := vpcListConfig{
		Region: types.StringNull(),
	}
	assert.True(t, nullConfig.Region.IsNull())
}

// TestVPCListResource_NewVPCListResource creates new instance.
func TestVPCListResource_NewVPCListResource(t *testing.T) {
	lr := NewVPCListResource()
	assert.NotNil(t, lr)
	assert.IsType(t, &VPCListResource{}, lr)
}

// TestVPCListResource_ToVPCModel tests the conversion helper.
func TestVPCListResource_ToVPCModel(t *testing.T) {
	now := time.Now()
	clientVPC := client.VPC{
		ID:        "vpc-1",
		Name:      "my-vpc",
		Region:    "us-east",
		Status:    "available",
		CRN:       "crn:vpc-1",
		CreatedAt: now,
	}

	model := toVPCModel(clientVPC)

	assert.Equal(t, "vpc-1", model.ID.ValueString())
	assert.Equal(t, "my-vpc", model.Name.ValueString())
	assert.Equal(t, "us-east", model.Region.ValueString())
	assert.Equal(t, "available", model.Status.ValueString())
	assert.Equal(t, "crn:vpc-1", model.CRN.ValueString())
	assert.Equal(t, now.String(), model.CreatedAt.ValueString())
}

// TestVPCListResource_ShouldIncludeVPC tests the filter helper.
func TestVPCListResource_ShouldIncludeVPC(t *testing.T) {
	vpc := client.VPC{
		ID:        "vpc-1",
		Name:      "my-vpc",
		Region:    "us-east",
		Status:    "available",
		CreatedAt: time.Now(),
	}

	// Test: no region filter (include all)
	config := vpcListConfig{Region: types.StringNull()}
	assert.True(t, shouldIncludeVPC(vpc, config))

	// Test: matching region (include)
	config = vpcListConfig{Region: types.StringValue("us-east")}
	assert.True(t, shouldIncludeVPC(vpc, config))

	// Test: non-matching region (exclude)
	config = vpcListConfig{Region: types.StringValue("us-west")}
	assert.False(t, shouldIncludeVPC(vpc, config))
}
