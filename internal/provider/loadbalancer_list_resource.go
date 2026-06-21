package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ list.ListResource = &LoadBalancerListResource{}
var _ list.ListResourceWithConfigure = &LoadBalancerListResource{}

type LoadBalancerListResource struct {
	client *client.Client
}

type loadbalancerListConfig struct {
	Protocol types.String `tfsdk:"protocol"`
}


func NewLoadBalancerListResource() list.ListResource {
	return &LoadBalancerListResource{}
}

func (r *LoadBalancerListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

func (r *LoadBalancerListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Query existing load balancers with optional filters.",
		Attributes: map[string]listschema.Attribute{
			"protocol": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter load balancers by protocol (HTTP, HTTPS, TCP, UDP).",
			},
		},
	}
}

func (r *LoadBalancerListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *LoadBalancerListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.client == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unconfigured Provider", "The provider hasn't been configured yet."),
		})
		return
	}

	var config loadbalancerListConfig
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	lbs, err := r.client.ListLoadBalancers()
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to List Load Balancers", err.Error()),
		})
		return
	}

	resp.Results = func(push func(list.ListResult) bool) {
		for _, lb := range lbs {
			if !config.Protocol.IsNull() && lb.Protocol != config.Protocol.ValueString() {
				continue
			}

			targetModels := make([]loadBalancerTargetModel, len(lb.Targets))
			for i, t := range lb.Targets {
				targetModels[i] = loadBalancerTargetModel{
					Type: types.StringValue(t.Type),
					ID:   types.StringValue(t.ID),
				}
			}
			targetsList, targetDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{"type": types.StringType, "id": types.StringType}}, targetModels)
			if targetDiags.HasError() {
				if !push(list.ListResult{Diagnostics: targetDiags}) {
					return
				}
				continue
			}

			lbModel := loadBalancerModel{
				ID:        types.StringValue(lb.ID),
				Name:      types.StringValue(lb.Name),
				Status:    types.StringValue(lb.Status),
				CRN:       types.StringValue(lb.CRN),
				Protocol:  types.StringValue(lb.Protocol),
				Port:      types.Int64Value(int64(lb.Port)),
				Targets:   targetsList,
				CreatedAt: types.StringValue(lb.CreatedAt.String()),
			}

			resourceSchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id":         resourceschema.StringAttribute{Computed: true},
					"name":       resourceschema.StringAttribute{Computed: true},
					"status":     resourceschema.StringAttribute{Computed: true},
					"crn":        resourceschema.StringAttribute{Computed: true},
					"protocol":   resourceschema.StringAttribute{Computed: true},
					"port":       resourceschema.Int64Attribute{Computed: true},
					"targets":    resourceschema.ListNestedAttribute{Computed: true, NestedObject: resourceschema.NestedAttributeObject{Attributes: map[string]resourceschema.Attribute{"type": resourceschema.StringAttribute{Computed: true}, "id": resourceschema.StringAttribute{Computed: true}}}},
					"created_at": resourceschema.StringAttribute{Computed: true},
				},
			}

			var val attr.Value
			diags := tfsdk.ValueFrom(ctx, lbModel, resourceSchema.Type(), &val)
			if diags.HasError() {
				if !push(list.ListResult{Diagnostics: diags}) {
					return
				}
				continue
			}

			tfVal, err := val.ToTerraformValue(ctx)
			if err != nil {
				if !push(list.ListResult{Diagnostics: diags}) {
					return
				}
				continue
			}

			identitySchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{Computed: true},
				},
			}

			identityVal, identityDiags := types.ObjectValue(
				map[string]attr.Type{"id": types.StringType},
				map[string]attr.Value{"id": lbModel.ID},
			)
			if identityDiags.HasError() {
				if !push(list.ListResult{Diagnostics: identityDiags}) {
					return
				}
				continue
			}

			identityTfVal, err := identityVal.ToTerraformValue(ctx)
			if err != nil {
				if !push(list.ListResult{Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("Identity Terraform Value Error", err.Error()),
				}}) {
					return
				}
				continue
			}

			if !push(list.ListResult{
				Identity: &tfsdk.ResourceIdentity{
					Raw:    identityTfVal,
					Schema: identitySchema,
				},
				Resource: &tfsdk.Resource{
					Raw:    tfVal,
					Schema: resourceSchema,
				},
				DisplayName: lbModel.Name.ValueString(),
			}) {
				return
			}
		}
	}
}

// toLoadBalancerModel converts a client.LoadBalancer to a loadbalancerModel for Terraform.

// toLoadBalancerModel converts a client.LoadBalancer to a loadBalancerModel for Terraform.
func toLoadBalancerModel(lb client.LoadBalancer) loadBalancerModel {
	return loadBalancerModel{
		ID:        types.StringValue(lb.ID),
		Name:      types.StringValue(lb.Name),
		Protocol:  types.StringValue(lb.Protocol),
		Status:    types.StringValue(lb.Status),
		CRN:       types.StringValue(lb.CRN),
		CreatedAt: types.StringValue(lb.CreatedAt.String()),
	}
}

// shouldIncludeLoadBalancer checks if a load balancer matches the filter config.
func shouldIncludeLoadBalancer(lb client.LoadBalancer, config loadbalancerListConfig) bool {
	if !config.Protocol.IsNull() && lb.Protocol != config.Protocol.ValueString() {
		return false
	}
	return true
}
