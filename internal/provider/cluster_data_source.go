package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &KubernetesClusterDataSource{}

type KubernetesClusterDataSource struct {
	client *client.Client
}

func NewKubernetesClusterDataSource() datasource.DataSource {
	return &KubernetesClusterDataSource{}
}

func (d *KubernetesClusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *KubernetesClusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a NullCloud Kubernetes cluster by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Kubernetes cluster.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"version": schema.StringAttribute{
				Computed: true,
			},
			"node_count": schema.Int64Attribute{
				Computed: true,
			},
			"subnet_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"crn": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *KubernetesClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *KubernetesClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data kubernetesClusterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, found, err := d.client.GetKubernetesCluster(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Kubernetes cluster", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Kubernetes cluster not found", fmt.Sprintf("Kubernetes cluster with ID %q not found", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(cluster.Name)
	data.Version = types.StringValue(cluster.Version)
	data.NodeCount = types.Int64Value(int64(cluster.NodeCount))
	data.Status = types.StringValue(cluster.Status)
	data.CRN = types.StringValue(cluster.CRN)
	data.SubnetIDs = stringsToList(cluster.SubnetIDs)
	data.CreatedAt = types.StringValue(cluster.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
