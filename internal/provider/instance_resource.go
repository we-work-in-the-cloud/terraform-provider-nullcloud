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

var _ resource.Resource = &InstanceResource{}

type InstanceResource struct {
	client *client.Client
}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

type instanceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	SubnetID  types.String `tfsdk:"subnet_id"`
	Profile   types.String `tfsdk:"profile"`
	Image     types.String `tfsdk:"image"`
	Status    types.String `tfsdk:"status"`
	VPCID     types.String `tfsdk:"vpc_id"`
	PrimaryIP types.String `tfsdk:"primary_ip"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NullCloud virtual server instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"profile": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *InstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data instanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := r.client.CreateInstance(
		data.Name.ValueString(),
		data.SubnetID.ValueString(),
		data.Profile.ValueString(),
		data.Image.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Instance", err.Error())
		return
	}

	data.ID = types.StringValue(inst.ID)
	data.Status = types.StringValue(inst.Status)
	data.VPCID = types.StringValue(inst.VPCID)
	data.PrimaryIP = types.StringValue(inst.PrimaryIP)
	data.Profile = types.StringValue(inst.Profile)
	data.Image = types.StringValue(inst.Image)
	data.CreatedAt = types.StringValue(inst.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data instanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst, found, err := r.client.GetInstance(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Instance", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(inst.Name)
	data.SubnetID = types.StringValue(inst.SubnetID)
	data.VPCID = types.StringValue(inst.VPCID)
	data.Status = types.StringValue(inst.Status)
	data.Profile = types.StringValue(inst.Profile)
	data.Image = types.StringValue(inst.Image)
	data.PrimaryIP = types.StringValue(inst.PrimaryIP)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable attributes require replace; Update is never called.
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data instanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteInstance(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting Instance", err.Error())
	}
}
