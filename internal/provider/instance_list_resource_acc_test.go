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

func TestInstanceListResource_Metadata(t *testing.T) {
	r := &InstanceListResource{}
	req := resource.MetadataRequest{ProviderTypeName: "nullcloud"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "nullcloud_instance", resp.TypeName)
}

func TestInstanceListResource_ListResourceConfigSchema(t *testing.T) {
	r := &InstanceListResource{}
	resp := &list.ListResourceSchemaResponse{}

	r.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "subnet_id")
	assert.Contains(t, resp.Schema.Attributes, "status")
}

func TestInstanceListResource_Configure(t *testing.T) {
	mockClient := &client.Client{}
	r := &InstanceListResource{}
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.Equal(t, mockClient, r.client)
}

func TestInstanceListResource_FilterLogic(t *testing.T) {
	instances := []client.Instance{
		{ID: "inst-1", Name: "inst-1", SubnetID: "subnet-1", Status: "running", CreatedAt: time.Now()},
		{ID: "inst-2", Name: "inst-2", SubnetID: "subnet-2", Status: "stopped", CreatedAt: time.Now()},
		{ID: "inst-3", Name: "inst-3", SubnetID: "subnet-1", Status: "running", CreatedAt: time.Now()},
	}

	// Test: no filter (all results)
	filtered := filterInstances(instances, "", "")
	assert.Equal(t, 3, len(filtered))

	// Test: filter by subnet
	filtered = filterInstances(instances, "subnet-1", "")
	assert.Equal(t, 2, len(filtered))

	// Test: filter by status
	filtered = filterInstances(instances, "", "running")
	assert.Equal(t, 2, len(filtered))

	// Test: filter by both
	filtered = filterInstances(instances, "subnet-1", "running")
	assert.Equal(t, 2, len(filtered))
}

func filterInstances(instances []client.Instance, subnetID, status string) []client.Instance {
	var result []client.Instance
	for _, inst := range instances {
		if (subnetID == "" || inst.SubnetID == subnetID) && (status == "" || inst.Status == status) {
			result = append(result, inst)
		}
	}
	return result
}

func TestInstanceListResource_ConfigModel(t *testing.T) {
	config := instanceListConfig{
		SubnetID: types.StringValue("subnet-1"),
		Status:   types.StringValue("running"),
	}

	assert.Equal(t, "subnet-1", config.SubnetID.ValueString())
	assert.Equal(t, "running", config.Status.ValueString())
}

func TestInstanceListResource_NewInstanceListResource(t *testing.T) {
	lr := NewInstanceListResource()
	assert.NotNil(t, lr)
	assert.IsType(t, &InstanceListResource{}, lr)
}

func TestInstanceListResource_ToInstanceModel(t *testing.T) {
	now := time.Now()
	clientInst := client.Instance{
		ID:        "inst-1",
		Name:      "my-instance",
		SubnetID:  "subnet-1",
		Status:    "running",
		CRN:       "crn:inst-1",
		CreatedAt: now,
	}

	model := toInstanceModel(clientInst)

	assert.Equal(t, "inst-1", model.ID.ValueString())
	assert.Equal(t, "subnet-1", model.SubnetID.ValueString())
	assert.Equal(t, "running", model.Status.ValueString())
}

func TestInstanceListResource_ShouldIncludeInstance(t *testing.T) {
	inst := client.Instance{
		ID:        "inst-1",
		SubnetID:  "subnet-1",
		Status:    "running",
		CreatedAt: time.Now(),
	}

	config := instanceListConfig{SubnetID: types.StringValue("subnet-1"), Status: types.StringValue("running")}
	assert.True(t, shouldIncludeInstance(inst, config))

	config = instanceListConfig{SubnetID: types.StringValue("subnet-2"), Status: types.StringNull()}
	assert.False(t, shouldIncludeInstance(inst, config))
}
