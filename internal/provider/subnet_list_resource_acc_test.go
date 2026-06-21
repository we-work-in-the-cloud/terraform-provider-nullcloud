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

func TestSubnetListResource_Metadata(t *testing.T) {
	r := &SubnetListResource{}
	req := resource.MetadataRequest{ProviderTypeName: "nullcloud"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "nullcloud_subnet", resp.TypeName)
}

func TestSubnetListResource_ListResourceConfigSchema(t *testing.T) {
	r := &SubnetListResource{}
	resp := &list.ListResourceSchemaResponse{}

	r.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "vpc_id")
	assert.Contains(t, resp.Schema.Attributes, "zone")
}

func TestSubnetListResource_Configure(t *testing.T) {
	mockClient := &client.Client{}
	r := &SubnetListResource{}
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.Equal(t, mockClient, r.client)
}

func TestSubnetListResource_FilterLogic(t *testing.T) {
	subnets := []client.Subnet{
		{ID: "subnet-1", Name: "subnet-1", VPCID: "vpc-1", Zone: "us-east-1", Status: "available", CreatedAt: time.Now()},
		{ID: "subnet-2", Name: "subnet-2", VPCID: "vpc-2", Zone: "us-east-1", Status: "available", CreatedAt: time.Now()},
		{ID: "subnet-3", Name: "subnet-3", VPCID: "vpc-1", Zone: "us-east-2", Status: "available", CreatedAt: time.Now()},
	}

	// Test: no filter (all results)
	filtered := filterSubnets(subnets, "", "")
	assert.Equal(t, 3, len(filtered))

	// Test: filter by VPCID
	filtered = filterSubnets(subnets, "vpc-1", "")
	assert.Equal(t, 2, len(filtered))

	// Test: filter by zone
	filtered = filterSubnets(subnets, "", "us-east-1")
	assert.Equal(t, 2, len(filtered))

	// Test: filter by both VPCID and zone
	filtered = filterSubnets(subnets, "vpc-1", "us-east-1")
	assert.Equal(t, 1, len(filtered))
	assert.Equal(t, "subnet-1", filtered[0].ID)
}

func filterSubnets(subnets []client.Subnet, vpcID, zone string) []client.Subnet {
	var result []client.Subnet
	for _, subnet := range subnets {
		if (vpcID == "" || subnet.VPCID == vpcID) && (zone == "" || subnet.Zone == zone) {
			result = append(result, subnet)
		}
	}
	return result
}

func TestSubnetListResource_ConfigModel(t *testing.T) {
	config := subnetListConfig{
		VPCID: types.StringValue("vpc-1"),
		Zone:  types.StringValue("us-east-1"),
	}

	assert.Equal(t, "vpc-1", config.VPCID.ValueString())
	assert.Equal(t, "us-east-1", config.Zone.ValueString())
	assert.False(t, config.VPCID.IsNull())
	assert.False(t, config.Zone.IsNull())
}

func TestSubnetListResource_NewSubnetListResource(t *testing.T) {
	lr := NewSubnetListResource()
	assert.NotNil(t, lr)
	assert.IsType(t, &SubnetListResource{}, lr)
}

func TestSubnetListResource_ToSubnetModel(t *testing.T) {
	now := time.Now()
	clientSubnet := client.Subnet{
		ID:        "subnet-1",
		Name:      "my-subnet",
		VPCID:     "vpc-1",
		Zone:      "us-east-1",
		Status:    "available",
		CRN:       "crn:subnet-1",
		CreatedAt: now,
	}

	model := toSubnetModel(clientSubnet)

	assert.Equal(t, "subnet-1", model.ID.ValueString())
	assert.Equal(t, "vpc-1", model.VPCID.ValueString())
	assert.Equal(t, "us-east-1", model.Zone.ValueString())
}

func TestSubnetListResource_ShouldIncludeSubnet(t *testing.T) {
	subnet := client.Subnet{
		ID:        "subnet-1",
		VPCID:     "vpc-1",
		Zone:      "us-east-1",
		Status:    "available",
		CreatedAt: time.Now(),
	}

	config := subnetListConfig{VPCID: types.StringValue("vpc-1"), Zone: types.StringValue("us-east-1")}
	assert.True(t, shouldIncludeSubnet(subnet, config))

	config = subnetListConfig{VPCID: types.StringValue("vpc-2"), Zone: types.StringNull()}
	assert.False(t, shouldIncludeSubnet(subnet, config))
}
