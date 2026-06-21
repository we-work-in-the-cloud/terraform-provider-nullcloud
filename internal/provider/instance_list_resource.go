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

var _ list.ListResource = &InstanceListResource{}
var _ list.ListResourceWithConfigure = &InstanceListResource{}

type InstanceListResource struct {
	client *client.Client
}

type instanceListConfig struct {
	SubnetID types.String `tfsdk:"subnet_id"`
	Status   types.String `tfsdk:"status"`
}


func NewInstanceListResource() list.ListResource {
	return &InstanceListResource{}
}

func (r *InstanceListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Query existing instances with optional filters.",
		Attributes: map[string]listschema.Attribute{
			"subnet_id": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter instances by subnet ID.",
			},
			"status": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter instances by status (e.g., running, stopped).",
			},
		},
	}
}

func (r *InstanceListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *InstanceListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.client == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unconfigured Provider", "The provider hasn't been configured yet."),
		})
		return
	}

	var config instanceListConfig
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	instances, err := r.client.ListInstances()
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to List Instances", err.Error()),
		})
		return
	}

	resp.Results = func(push func(list.ListResult) bool) {
		for _, instance := range instances {
			if !config.SubnetID.IsNull() && instance.SubnetID != config.SubnetID.ValueString() {
				continue
			}
			if !config.Status.IsNull() && instance.Status != config.Status.ValueString() {
				continue
			}

			instanceModel := instanceModel{
				ID:        types.StringValue(instance.ID),
				Name:      types.StringValue(instance.Name),
				SubnetID:  types.StringValue(instance.SubnetID),
				Profile:   types.StringValue(instance.Profile),
				Image:     types.StringValue(instance.Image),
				Status:    types.StringValue(instance.Status),
				CRN:       types.StringValue(instance.CRN),
				PrimaryIP: types.StringValue(instance.PrimaryIP),
				CreatedAt: types.StringValue(instance.CreatedAt.String()),
			}

			resourceSchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id":         resourceschema.StringAttribute{Computed: true},
					"name":       resourceschema.StringAttribute{Computed: true},
					"subnet_id":  resourceschema.StringAttribute{Computed: true},
					"profile":    resourceschema.StringAttribute{Computed: true},
					"image":      resourceschema.StringAttribute{Computed: true},
					"status":     resourceschema.StringAttribute{Computed: true},
					"crn":        resourceschema.StringAttribute{Computed: true},
					"primary_ip": resourceschema.StringAttribute{Computed: true},
					"created_at": resourceschema.StringAttribute{Computed: true},
				},
			}

			identitySchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{Computed: true},
				},
			}

			var val attr.Value
			diags := tfsdk.ValueFrom(ctx, instanceModel, resourceSchema.Type(), &val)
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

			identityVal, identityDiags := types.ObjectValue(
				map[string]attr.Type{"id": types.StringType},
				map[string]attr.Value{"id": instanceModel.ID},
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
				DisplayName: instanceModel.Name.ValueString(),
			}) {
				return
			}
		}
	}
}

// toInstanceModel converts a client.Instance to an instanceModel for Terraform.
func toInstanceModel(instance client.Instance) instanceModel {
	return instanceModel{
		ID:        types.StringValue(instance.ID),
		Name:      types.StringValue(instance.Name),
		SubnetID:  types.StringValue(instance.SubnetID),
		Status:    types.StringValue(instance.Status),
		CRN:       types.StringValue(instance.CRN),
		CreatedAt: types.StringValue(instance.CreatedAt.String()),
	}
}

// shouldIncludeInstance checks if an instance matches the filter config.
func shouldIncludeInstance(instance client.Instance, config instanceListConfig) bool {
	if !config.SubnetID.IsNull() && instance.SubnetID != config.SubnetID.ValueString() {
		return false
	}
	if !config.Status.IsNull() && instance.Status != config.Status.ValueString() {
		return false
	}
	return true
}
