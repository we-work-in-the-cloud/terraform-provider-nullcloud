package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &SubnetDataSource{}

type SubnetDataSource struct {
	client *client.Client
}

func NewSubnetDataSource() datasource.DataSource {
	return &SubnetDataSource{}
}

func (d *SubnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *SubnetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a NullCloud Subnet by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the subnet.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"vpc_id": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"crn": schema.StringAttribute{
				Computed: true,
			},
			"cidr_block": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *SubnetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subnetModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, found, err := d.client.GetSubnet(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Subnet", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Subnet not found", fmt.Sprintf("Subnet with ID %q not found", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(sub.Name)
	data.VPCID = types.StringValue(sub.VPCID)
	data.Status = types.StringValue(sub.Status)
	data.CRN = types.StringValue(sub.CRN)
	data.CIDRBlock = types.StringValue(sub.CIDRBlock)
	data.CreatedAt = types.StringValue(sub.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
