package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &LoadBalancerDataSource{}

type LoadBalancerDataSource struct {
	client *client.Client
}

func NewLoadBalancerDataSource() datasource.DataSource {
	return &LoadBalancerDataSource{}
}

func (d *LoadBalancerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

func (d *LoadBalancerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a NullCloud load balancer by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the load balancer.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"protocol": schema.StringAttribute{
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Computed: true,
			},
			"targets": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of targets (cluster or vsi) this load balancer routes traffic to.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{Computed: true},
						"id":   schema.StringAttribute{Computed: true},
					},
				},
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

func (d *LoadBalancerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *LoadBalancerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, found, err := d.client.GetLoadBalancer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading load balancer", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Load balancer not found", fmt.Sprintf("Load balancer with ID %q not found", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(lb.Name)
	data.Status = types.StringValue(lb.Status)
	data.CRN = types.StringValue(lb.CRN)
	data.Protocol = types.StringValue(lb.Protocol)
	data.Port = types.Int64Value(int64(lb.Port))
	data.Targets = lbTargetsList(lb.Targets)
	data.CreatedAt = types.StringValue(lb.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
