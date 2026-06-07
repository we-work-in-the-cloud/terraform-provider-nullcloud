package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &InstanceDataSource{}

type InstanceDataSource struct {
	client *client.Client
}

func NewInstanceDataSource() datasource.DataSource {
	return &InstanceDataSource{}
}

func (d *InstanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (d *InstanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a NullCloud virtual server instance by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the instance.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"subnet_id": schema.StringAttribute{
				Computed: true,
			},
			"profile": schema.StringAttribute{
				Computed: true,
			},
			"image": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"crn": schema.StringAttribute{
				Computed: true,
			},
			"primary_ip": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *InstanceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *InstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data instanceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst, found, err := d.client.GetInstance(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Instance", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Instance not found", fmt.Sprintf("Instance with ID %q not found", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(inst.Name)
	data.SubnetID = types.StringValue(inst.SubnetID)
	data.Profile = types.StringValue(inst.Profile)
	data.Image = types.StringValue(inst.Image)
	data.Status = types.StringValue(inst.Status)
	data.CRN = types.StringValue(inst.CRN)
	data.PrimaryIP = types.StringValue(inst.PrimaryIP)
	data.CreatedAt = types.StringValue(inst.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
