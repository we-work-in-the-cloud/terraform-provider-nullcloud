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

var _ list.ListResource = &SubnetListResource{}
var _ list.ListResourceWithConfigure = &SubnetListResource{}

type SubnetListResource struct {
	client *client.Client
}

type subnetListConfig struct {
	VPCID types.String `tfsdk:"vpc_id"`
	Zone  types.String `tfsdk:"zone"`
}


func NewSubnetListResource() list.ListResource {
	return &SubnetListResource{}
}

func (r *SubnetListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Query existing subnets with optional filters.",
		Attributes: map[string]listschema.Attribute{
			"vpc_id": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter subnets by VPC ID.",
			},
			"zone": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter subnets by zone.",
			},
		},
	}
}

func (r *SubnetListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *SubnetListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.client == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unconfigured Provider", "The provider hasn't been configured yet."),
		})
		return
	}

	var config subnetListConfig
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	subnets, err := r.client.ListSubnets()
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to List Subnets", err.Error()),
		})
		return
	}

	resp.Results = func(push func(list.ListResult) bool) {
		for _, subnet := range subnets {
			if !config.VPCID.IsNull() && subnet.VPCID != config.VPCID.ValueString() {
				continue
			}
			if !config.Zone.IsNull() && subnet.Zone != config.Zone.ValueString() {
				continue
			}

			subnetModel := subnetModel{
				ID:        types.StringValue(subnet.ID),
				Name:      types.StringValue(subnet.Name),
				VPCID:     types.StringValue(subnet.VPCID),
				Zone:      types.StringValue(subnet.Zone),
				Status:    types.StringValue(subnet.Status),
				CRN:       types.StringValue(subnet.CRN),
				CIDRBlock: types.StringValue(subnet.CIDRBlock),
				CreatedAt: types.StringValue(subnet.CreatedAt.String()),
			}

			resourceSchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id":         resourceschema.StringAttribute{Computed: true},
					"name":       resourceschema.StringAttribute{Computed: true},
					"vpc_id":     resourceschema.StringAttribute{Computed: true},
					"zone":       resourceschema.StringAttribute{Computed: true},
					"status":     resourceschema.StringAttribute{Computed: true},
					"crn":        resourceschema.StringAttribute{Computed: true},
					"cidr_block": resourceschema.StringAttribute{Computed: true},
					"created_at": resourceschema.StringAttribute{Computed: true},
				},
			}

			identitySchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{Computed: true},
				},
			}

			var val attr.Value
			diags := tfsdk.ValueFrom(ctx, subnetModel, resourceSchema.Type(), &val)
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
				map[string]attr.Value{"id": subnetModel.ID},
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
				DisplayName: subnetModel.Name.ValueString(),
			}) {
				return
			}
		}
	}
}

// toSubnetModel converts a client.Subnet to a subnetModel for Terraform.
func toSubnetModel(subnet client.Subnet) subnetModel {
	return subnetModel{
		ID:        types.StringValue(subnet.ID),
		Name:      types.StringValue(subnet.Name),
		VPCID:     types.StringValue(subnet.VPCID),
		Zone:      types.StringValue(subnet.Zone),
		Status:    types.StringValue(subnet.Status),
		CRN:       types.StringValue(subnet.CRN),
		CreatedAt: types.StringValue(subnet.CreatedAt.String()),
	}
}

// shouldIncludeSubnet checks if a subnet matches the filter config.
func shouldIncludeSubnet(subnet client.Subnet, config subnetListConfig) bool {
	if !config.VPCID.IsNull() && subnet.VPCID != config.VPCID.ValueString() {
		return false
	}
	if !config.Zone.IsNull() && subnet.Zone != config.Zone.ValueString() {
		return false
	}
	return true
}
