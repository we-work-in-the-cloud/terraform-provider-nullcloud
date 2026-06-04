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

var _ resource.Resource = &VPCResource{}

type VPCResource struct {
	client *client.Client
}

func NewVPCResource() resource.Resource {
	return &VPCResource{}
}

type vpcModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Region    types.String `tfsdk:"region"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *VPCResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *VPCResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NullCloud VPC.",
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
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Region where the VPC is created. Defaults to the provider region (us-east).",
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
			"crn": schema.StringAttribute{
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

func (r *VPCResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *VPCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vpcModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = r.client.DefaultRegion
	}

	vpc, err := r.client.CreateVPC(data.Name.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Error creating VPC", err.Error())
		return
	}

	data.ID = types.StringValue(vpc.ID)
	data.Region = types.StringValue(vpc.Region)
	data.Status = types.StringValue(vpc.Status)
	data.CRN = types.StringValue(vpc.CRN)
	data.CreatedAt = types.StringValue(vpc.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vpcModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpc, found, err := r.client.GetVPC(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading VPC", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(vpc.Name)
	data.Region = types.StringValue(vpc.Region)
	data.Status = types.StringValue(vpc.Status)
	data.CRN = types.StringValue(vpc.CRN)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable attributes require replace; Update is never called.
}

func (r *VPCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vpcModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVPC(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting VPC", err.Error())
	}
}
