package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &VPCDataSource{}

type VPCDataSource struct {
	client *client.Client
}

func NewVPCDataSource() datasource.DataSource {
	return &VPCDataSource{}
}

func (d *VPCDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (d *VPCDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a NullCloud VPC by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the VPC.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"region": schema.StringAttribute{
				Computed: true,
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

func (d *VPCDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *VPCDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vpcModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpc, found, err := d.client.GetVPC(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading VPC", err.Error())
		return
	}
	if !found {
		reportNotFoundDataSource(resp, "VPC", data.ID.ValueString())
		return
	}

	data.Name = types.StringValue(vpc.Name)
	data.Region = types.StringValue(vpc.Region)
	data.Status = types.StringValue(vpc.Status)
	data.CRN = types.StringValue(vpc.CRN)
	data.CreatedAt = types.StringValue(vpc.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
