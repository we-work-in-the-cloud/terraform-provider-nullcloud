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

var _ list.ListResource = &VPCListResource{}
var _ list.ListResourceWithConfigure = &VPCListResource{}

type VPCListResource struct {
	client *client.Client
}

type vpcListConfig struct {
	Region types.String `tfsdk:"region"`
}

func NewVPCListResource() list.ListResource {
	return &VPCListResource{}
}

// toVPCModel converts a client.VPC to a vpcModel for Terraform.
func toVPCModel(vpc client.VPC) vpcModel {
	return vpcModel{
		ID:        types.StringValue(vpc.ID),
		Name:      types.StringValue(vpc.Name),
		Region:    types.StringValue(vpc.Region),
		Status:    types.StringValue(vpc.Status),
		CRN:       types.StringValue(vpc.CRN),
		CreatedAt: types.StringValue(vpc.CreatedAt.String()),
	}
}

// shouldIncludeVPC checks if a VPC matches the filter config.
func shouldIncludeVPC(vpc client.VPC, config vpcListConfig) bool {
	if !config.Region.IsNull() && vpc.Region != config.Region.ValueString() {
		return false
	}
	return true
}

func (r *VPCListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *VPCListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Query existing VPCs with optional filters.",
		Attributes: map[string]listschema.Attribute{
			"region": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter VPCs by region.",
			},
		},
	}
}

func (r *VPCListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *VPCListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.client == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unconfigured Provider", "The provider hasn't been configured yet."),
		})
		return
	}

	var config vpcListConfig
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	vpcs, err := r.client.ListVPCs()
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to List VPCs", err.Error()),
		})
		return
	}

	resp.Results = func(push func(list.ListResult) bool) {
		for _, vpc := range vpcs {
			if !shouldIncludeVPC(vpc, config) {
				continue
			}

			vpcModel := toVPCModel(vpc)

			resourceSchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{
						Computed: true,
					},
					"name": resourceschema.StringAttribute{
						Computed: true,
					},
					"region": resourceschema.StringAttribute{
						Computed: true,
					},
					"status": resourceschema.StringAttribute{
						Computed: true,
					},
					"crn": resourceschema.StringAttribute{
						Computed: true,
					},
					"created_at": resourceschema.StringAttribute{
						Computed: true,
					},
				},
			}

			identitySchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{
						Computed: true,
					},
				},
			}

			var val attr.Value
			diags := tfsdk.ValueFrom(ctx, vpcModel, resourceSchema.Type(), &val)
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
				map[string]attr.Value{"id": vpcModel.ID},
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
				DisplayName: vpcModel.Name.ValueString(),
			}) {
				return
			}
		}
	}
}
