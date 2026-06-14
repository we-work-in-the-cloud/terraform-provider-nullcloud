package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ resource.Resource = &SubnetResource{}

type SubnetResource struct {
	client *client.Client
}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

func (r *SubnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NullCloud Subnet.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"vpc_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": schema.StringAttribute{
				Required:    true,
				Description: "Availability zone for the subnet (e.g. us-east-1). Must be a zone within the VPC's region.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"crn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cidr_block": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SubnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data subnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.CreateSubnet(data.Name.ValueString(), data.VPCID.ValueString(), data.Zone.ValueString(), data.CIDRBlock.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Subnet", err.Error())
		return
	}

	data.ID = types.StringValue(sub.ID)
	data.Zone = types.StringValue(sub.Zone)
	data.Status = types.StringValue(sub.Status)
	data.CRN = types.StringValue(sub.CRN)
	data.CIDRBlock = types.StringValue(sub.CIDRBlock)
	data.CreatedAt = types.StringValue(sub.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data subnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, found, err := r.client.GetSubnet(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Subnet", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(sub.Name)
	data.VPCID = types.StringValue(sub.VPCID)
	data.Zone = types.StringValue(sub.Zone)
	data.Status = types.StringValue(sub.Status)
	data.CRN = types.StringValue(sub.CRN)
	data.CIDRBlock = types.StringValue(sub.CIDRBlock)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data subnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.UpdateSubnet(data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating Subnet", err.Error())
		return
	}

	data.Name = types.StringValue(sub.Name)
	data.Status = types.StringValue(sub.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data subnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSubnet(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting Subnet", err.Error())
	}
}
