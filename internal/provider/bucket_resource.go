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

var _ resource.Resource = &BucketResource{}

type BucketResource struct {
	client *client.Client
}

func NewBucketResource() resource.Resource {
	return &BucketResource{}
}

type bucketModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Region    types.String `tfsdk:"region"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *BucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (r *BucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NullCloud object storage bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Region where the bucket is created. Defaults to the provider region (us-east).",
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

func (r *BucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = r.client.DefaultRegion
	}

	b, err := r.client.CreateBucket(data.Name.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket", err.Error())
		return
	}

	data.ID = types.StringValue(b.ID)
	data.Region = types.StringValue(b.Region)
	data.Status = types.StringValue(b.Status)
	data.CRN = types.StringValue(b.CRN)
	data.CreatedAt = types.StringValue(b.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, found, err := r.client.GetBucket(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(b.Name)
	data.Region = types.StringValue(b.Region)
	data.Status = types.StringValue(b.Status)
	data.CRN = types.StringValue(b.CRN)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BucketResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable attributes require replace; Update is never called.
}

func (r *BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBucket(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting bucket", err.Error())
	}
}
