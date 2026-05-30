package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ resource.Resource = &KubernetesClusterResource{}

type KubernetesClusterResource struct {
	client *client.Client
}

func NewKubernetesClusterResource() resource.Resource {
	return &KubernetesClusterResource{}
}

type kubernetesClusterModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Version   types.String `tfsdk:"version"`
	NodeCount types.Int64  `tfsdk:"node_count"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *KubernetesClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *KubernetesClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NullCloud Kubernetes cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Kubernetes cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "Kubernetes version (e.g. 1.30).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"node_count": schema.Int64Attribute{
				Required:    true,
				Description: "Number of worker nodes (minimum 1).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
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

func (r *KubernetesClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *KubernetesClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data kubernetesClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.CreateKubernetesCluster(
		data.Name.ValueString(),
		data.Version.ValueString(),
		int(data.NodeCount.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Kubernetes cluster", err.Error())
		return
	}

	data.ID = types.StringValue(cluster.ID)
	data.Status = types.StringValue(cluster.Status)
	data.CRN = types.StringValue(cluster.CRN)
	data.CreatedAt = types.StringValue(cluster.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, found, err := r.client.GetKubernetesCluster(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Kubernetes cluster", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(cluster.Name)
	data.Version = types.StringValue(cluster.Version)
	data.NodeCount = types.Int64Value(int64(cluster.NodeCount))
	data.Status = types.StringValue(cluster.Status)
	data.CRN = types.StringValue(cluster.CRN)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesClusterResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable attributes require replace; Update is never called.
}

func (r *KubernetesClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data kubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteKubernetesCluster(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting Kubernetes cluster", err.Error())
	}
}
