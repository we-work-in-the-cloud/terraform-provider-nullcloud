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

func TestLoadBalancerListResource_Metadata(t *testing.T) {
	r := &LoadBalancerListResource{}
	req := resource.MetadataRequest{ProviderTypeName: "nullcloud"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "nullcloud_loadbalancer", resp.TypeName)
}

func TestLoadBalancerListResource_ListResourceConfigSchema(t *testing.T) {
	r := &LoadBalancerListResource{}
	resp := &list.ListResourceSchemaResponse{}

	r.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "protocol")
}

func TestLoadBalancerListResource_Configure(t *testing.T) {
	mockClient := &client.Client{}
	r := &LoadBalancerListResource{}
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.Equal(t, mockClient, r.client)
}

func TestLoadBalancerListResource_FilterLogic(t *testing.T) {
	lbs := []client.LoadBalancer{
		{ID: "lb-1", Name: "lb-1", Protocol: "http", Port: 80, Status: "active", CreatedAt: time.Now()},
		{ID: "lb-2", Name: "lb-2", Protocol: "https", Port: 443, Status: "active", CreatedAt: time.Now()},
		{ID: "lb-3", Name: "lb-3", Protocol: "http", Port: 8080, Status: "active", CreatedAt: time.Now()},
	}

	// Test: no filter (all results)
	filtered := filterLoadBalancers(lbs, "")
	assert.Equal(t, 3, len(filtered))

	// Test: filter by protocol
	filtered = filterLoadBalancers(lbs, "http")
	assert.Equal(t, 2, len(filtered))

	filtered = filterLoadBalancers(lbs, "https")
	assert.Equal(t, 1, len(filtered))
}

func filterLoadBalancers(lbs []client.LoadBalancer, protocol string) []client.LoadBalancer {
	var result []client.LoadBalancer
	for _, lb := range lbs {
		if protocol == "" || lb.Protocol == protocol {
			result = append(result, lb)
		}
	}
	return result
}

func TestLoadBalancerListResource_ConfigModel(t *testing.T) {
	config := loadbalancerListConfig{
		Protocol: types.StringValue("http"),
	}

	assert.Equal(t, "http", config.Protocol.ValueString())
}

func TestLoadBalancerListResource_NewLoadBalancerListResource(t *testing.T) {
	lr := NewLoadBalancerListResource()
	assert.NotNil(t, lr)
	assert.IsType(t, &LoadBalancerListResource{}, lr)
}

func TestLoadBalancerListResource_ToLoadBalancerModel(t *testing.T) {
	now := time.Now()
	clientLB := client.LoadBalancer{
		ID:        "lb-1",
		Name:      "my-lb",
		Protocol:  "http",
		Status:    "active",
		CRN:       "crn:lb-1",
		CreatedAt: now,
	}

	model := toLoadBalancerModel(clientLB)

	assert.Equal(t, "lb-1", model.ID.ValueString())
	assert.Equal(t, "http", model.Protocol.ValueString())
	assert.Equal(t, "active", model.Status.ValueString())
}

func TestLoadBalancerListResource_ShouldIncludeLoadBalancer(t *testing.T) {
	lb := client.LoadBalancer{
		ID:        "lb-1",
		Protocol:  "http",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	config := loadbalancerListConfig{Protocol: types.StringValue("http")}
	assert.True(t, shouldIncludeLoadBalancer(lb, config))

	config = loadbalancerListConfig{Protocol: types.StringValue("https")}
	assert.False(t, shouldIncludeLoadBalancer(lb, config))
}
