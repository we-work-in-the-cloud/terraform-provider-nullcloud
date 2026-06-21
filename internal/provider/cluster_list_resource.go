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

var _ list.ListResource = &KubernetesClusterListResource{}
var _ list.ListResourceWithConfigure = &KubernetesClusterListResource{}

type KubernetesClusterListResource struct {
	client *client.Client
}

type kubernetesClusterListConfig struct {
	Version types.String `tfsdk:"version"`
}


func NewKubernetesClusterListResource() list.ListResource {
	return &KubernetesClusterListResource{}
}

func (r *KubernetesClusterListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *KubernetesClusterListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Query existing Kubernetes clusters with optional filters.",
		Attributes: map[string]listschema.Attribute{
			"version": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter clusters by version.",
			},
		},
	}
}

func (r *KubernetesClusterListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *KubernetesClusterListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.client == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unconfigured Provider", "The provider hasn't been configured yet."),
		})
		return
	}

	var config kubernetesClusterListConfig
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	clusters, err := r.client.ListKubernetesClusters()
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to List Kubernetes Clusters", err.Error()),
		})
		return
	}

	resp.Results = func(push func(list.ListResult) bool) {
		for _, cluster := range clusters {
			if !config.Version.IsNull() && cluster.Version != config.Version.ValueString() {
				continue
			}

			subnetIDValues := make([]attr.Value, len(cluster.SubnetIDs))
			for i, id := range cluster.SubnetIDs {
				subnetIDValues[i] = types.StringValue(id)
			}
			subnetIDsList, subnetDiags := types.ListValue(types.StringType, subnetIDValues)
			if subnetDiags.HasError() {
				if !push(list.ListResult{Diagnostics: subnetDiags}) {
					return
				}
				continue
			}

			clusterModel := kubernetesClusterModel{
				ID:        types.StringValue(cluster.ID),
				Name:      types.StringValue(cluster.Name),
				Status:    types.StringValue(cluster.Status),
				CRN:       types.StringValue(cluster.CRN),
				Version:   types.StringValue(cluster.Version),
				NodeCount: types.Int64Value(int64(cluster.NodeCount)),
				SubnetIDs: subnetIDsList,
				CreatedAt: types.StringValue(cluster.CreatedAt.String()),
			}

			resourceSchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id":         resourceschema.StringAttribute{Computed: true},
					"name":       resourceschema.StringAttribute{Computed: true},
					"status":     resourceschema.StringAttribute{Computed: true},
					"crn":        resourceschema.StringAttribute{Computed: true},
					"version":    resourceschema.StringAttribute{Computed: true},
					"node_count": resourceschema.Int64Attribute{Computed: true},
					"subnet_ids": resourceschema.ListAttribute{Computed: true, ElementType: types.StringType},
					"created_at": resourceschema.StringAttribute{Computed: true},
				},
			}

			identitySchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{Computed: true},
				},
			}

			var val attr.Value
			diags := tfsdk.ValueFrom(ctx, clusterModel, resourceSchema.Type(), &val)
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
				map[string]attr.Value{"id": clusterModel.ID},
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
				DisplayName: clusterModel.Name.ValueString(),
			}) {
				return
			}
		}
	}
}

// toKubernetesClusterModel converts a client.KubernetesCluster to a kubernetesClusterModel for Terraform.
func toKubernetesClusterModel(cluster client.KubernetesCluster) kubernetesClusterModel {
	return kubernetesClusterModel{
		ID:        types.StringValue(cluster.ID),
		Name:      types.StringValue(cluster.Name),
		Version:   types.StringValue(cluster.Version),
		Status:    types.StringValue(cluster.Status),
		CRN:       types.StringValue(cluster.CRN),
		CreatedAt: types.StringValue(cluster.CreatedAt.String()),
	}
}

// shouldIncludeKubernetesCluster checks if a cluster matches the filter config.
func shouldIncludeKubernetesCluster(cluster client.KubernetesCluster, config kubernetesClusterListConfig) bool {
	if !config.Version.IsNull() && cluster.Version != config.Version.ValueString() {
		return false
	}
	return true
}
