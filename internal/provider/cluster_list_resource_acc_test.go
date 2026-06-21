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

func TestKubernetesClusterListResource_Metadata(t *testing.T) {
	r := &KubernetesClusterListResource{}
	req := resource.MetadataRequest{ProviderTypeName: "nullcloud"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "nullcloud_cluster", resp.TypeName)
}

func TestKubernetesClusterListResource_ListResourceConfigSchema(t *testing.T) {
	r := &KubernetesClusterListResource{}
	resp := &list.ListResourceSchemaResponse{}

	r.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "version")
}

func TestKubernetesClusterListResource_Configure(t *testing.T) {
	mockClient := &client.Client{}
	r := &KubernetesClusterListResource{}
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.Equal(t, mockClient, r.client)
}

func TestKubernetesClusterListResource_FilterLogic(t *testing.T) {
	clusters := []client.KubernetesCluster{
		{ID: "cluster-1", Name: "cluster-1", Version: "1.24", Status: "active", CreatedAt: time.Now()},
		{ID: "cluster-2", Name: "cluster-2", Version: "1.25", Status: "active", CreatedAt: time.Now()},
		{ID: "cluster-3", Name: "cluster-3", Version: "1.24", Status: "active", CreatedAt: time.Now()},
	}

	// Test: no filter (all results)
	filtered := filterClusters(clusters, "")
	assert.Equal(t, 3, len(filtered))

	// Test: filter by version
	filtered = filterClusters(clusters, "1.24")
	assert.Equal(t, 2, len(filtered))

	filtered = filterClusters(clusters, "1.25")
	assert.Equal(t, 1, len(filtered))
}

func filterClusters(clusters []client.KubernetesCluster, version string) []client.KubernetesCluster {
	var result []client.KubernetesCluster
	for _, cluster := range clusters {
		if version == "" || cluster.Version == version {
			result = append(result, cluster)
		}
	}
	return result
}

func TestKubernetesClusterListResource_ConfigModel(t *testing.T) {
	config := kubernetesClusterListConfig{
		Version: types.StringValue("1.24"),
	}

	assert.Equal(t, "1.24", config.Version.ValueString())
}

func TestKubernetesClusterListResource_NewKubernetesClusterListResource(t *testing.T) {
	lr := NewKubernetesClusterListResource()
	assert.NotNil(t, lr)
	assert.IsType(t, &KubernetesClusterListResource{}, lr)
}

func TestKubernetesClusterListResource_ToKubernetesClusterModel(t *testing.T) {
	now := time.Now()
	clientCluster := client.KubernetesCluster{
		ID:        "cluster-1",
		Name:      "my-cluster",
		Version:   "1.24",
		Status:    "active",
		CRN:       "crn:cluster-1",
		CreatedAt: now,
	}

	model := toKubernetesClusterModel(clientCluster)

	assert.Equal(t, "cluster-1", model.ID.ValueString())
	assert.Equal(t, "1.24", model.Version.ValueString())
	assert.Equal(t, "active", model.Status.ValueString())
}

func TestKubernetesClusterListResource_ShouldIncludeKubernetesCluster(t *testing.T) {
	cluster := client.KubernetesCluster{
		ID:        "cluster-1",
		Version:   "1.24",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	config := kubernetesClusterListConfig{Version: types.StringValue("1.24")}
	assert.True(t, shouldIncludeKubernetesCluster(cluster, config))

	config = kubernetesClusterListConfig{Version: types.StringValue("1.25")}
	assert.False(t, shouldIncludeKubernetesCluster(cluster, config))
}
