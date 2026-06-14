package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ resource.Resource = &LoadBalancerResource{}

// lbTargetAttrTypes defines the object shape for a load balancer target.
var lbTargetAttrTypes = map[string]attr.Type{
	"type": types.StringType,
	"id":   types.StringType,
}

type LoadBalancerResource struct {
	client *client.Client
}

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

func (r *LoadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

func (r *LoadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NullCloud load balancer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the load balancer.",
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "Protocol for the load balancer. Must be tcp, http, or https.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "Port the load balancer listens on (1-65535).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"targets": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of targets (cluster or vsi) to route traffic to.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Target type. Must be cluster or vsi.",
						},
						"id": schema.StringAttribute{
							Required:    true,
							Description: "ID of the target resource.",
						},
					},
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

func (r *LoadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *LoadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var targets []client.LoadBalancerTarget
	if !data.Targets.IsNull() {
		var targetModels []loadBalancerTargetModel
		resp.Diagnostics.Append(data.Targets.ElementsAs(ctx, &targetModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		targets = make([]client.LoadBalancerTarget, len(targetModels))
		for i, t := range targetModels {
			targets[i] = client.LoadBalancerTarget{Type: t.Type.ValueString(), ID: t.ID.ValueString()}
		}
	}

	lb, err := r.client.CreateLoadBalancer(
		data.Name.ValueString(),
		data.Protocol.ValueString(),
		int(data.Port.ValueInt64()),
		targets,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating load balancer", err.Error())
		return
	}

	data.ID = types.StringValue(lb.ID)
	data.Status = types.StringValue(lb.Status)
	data.CRN = types.StringValue(lb.CRN)
	data.Targets = lbTargetsList(lb.Targets)
	data.CreatedAt = types.StringValue(lb.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, found, err := r.client.GetLoadBalancer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading load balancer", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(lb.Name)
	data.Status = types.StringValue(lb.Status)
	data.CRN = types.StringValue(lb.CRN)
	data.Protocol = types.StringValue(lb.Protocol)
	data.Port = types.Int64Value(int64(lb.Port))
	data.Targets = lbTargetsList(lb.Targets)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var targets []client.LoadBalancerTarget
	if !data.Targets.IsNull() {
		var targetModels []loadBalancerTargetModel
		resp.Diagnostics.Append(data.Targets.ElementsAs(ctx, &targetModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		targets = make([]client.LoadBalancerTarget, len(targetModels))
		for i, t := range targetModels {
			targets[i] = client.LoadBalancerTarget{Type: t.Type.ValueString(), ID: t.ID.ValueString()}
		}
	}

	lb, err := r.client.UpdateLoadBalancer(data.ID.ValueString(), data.Name.ValueString(), targets)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Load Balancer", err.Error())
		return
	}

	data.Name = types.StringValue(lb.Name)
	data.Status = types.StringValue(lb.Status)
	data.Targets = lbTargetsList(lb.Targets)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data loadBalancerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLoadBalancer(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting load balancer", err.Error())
	}
}

// lbTargetsList converts API targets to a types.List. Returns null list when targets is nil/empty.
func lbTargetsList(targets []client.LoadBalancerTarget) types.List {
	objType := types.ObjectType{AttrTypes: lbTargetAttrTypes}
	if len(targets) == 0 {
		return types.ListNull(objType)
	}
	elems := make([]attr.Value, len(targets))
	for i, t := range targets {
		obj, _ := types.ObjectValue(lbTargetAttrTypes, map[string]attr.Value{
			"type": types.StringValue(t.Type),
			"id":   types.StringValue(t.ID),
		})
		elems[i] = obj
	}
	result, _ := types.ListValue(objType, elems)
	return result
}
