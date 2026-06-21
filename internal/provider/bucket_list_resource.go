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

var _ list.ListResource = &BucketListResource{}
var _ list.ListResourceWithConfigure = &BucketListResource{}

type BucketListResource struct {
	client *client.Client
}

type bucketListConfig struct {
	Region types.String `tfsdk:"region"`
}


func NewBucketListResource() list.ListResource {
	return &BucketListResource{}
}

// toBucketModel converts a client.Bucket to a bucketModel for Terraform.
func toBucketModel(bucket client.Bucket) bucketModel {
	return bucketModel{
		ID:        types.StringValue(bucket.ID),
		Name:      types.StringValue(bucket.Name),
		Status:    types.StringValue(bucket.Status),
		CRN:       types.StringValue(bucket.CRN),
		Region:    types.StringValue(bucket.Region),
		CreatedAt: types.StringValue(bucket.CreatedAt.String()),
	}
}

// shouldIncludeBucket checks if a bucket matches the filter config.
func shouldIncludeBucket(bucket client.Bucket, config bucketListConfig) bool {
	if !config.Region.IsNull() && bucket.Region != config.Region.ValueString() {
		return false
	}
	return true
}

func (r *BucketListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (r *BucketListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Query existing buckets with optional filters.",
		Attributes: map[string]listschema.Attribute{
			"region": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter buckets by region.",
			},
		},
	}
}

func (r *BucketListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *BucketListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.client == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unconfigured Provider", "The provider hasn't been configured yet."),
		})
		return
	}

	var config bucketListConfig
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	buckets, err := r.client.ListBuckets()
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to List Buckets", err.Error()),
		})
		return
	}

	resp.Results = func(push func(list.ListResult) bool) {
		for _, bucket := range buckets {
			if !shouldIncludeBucket(bucket, config) {
				continue
			}

			bucketModel := toBucketModel(bucket)

			resourceSchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id":         resourceschema.StringAttribute{Computed: true},
					"name":       resourceschema.StringAttribute{Computed: true},
					"status":     resourceschema.StringAttribute{Computed: true},
					"crn":        resourceschema.StringAttribute{Computed: true},
					"region":     resourceschema.StringAttribute{Computed: true},
					"created_at": resourceschema.StringAttribute{Computed: true},
				},
			}

			identitySchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{Computed: true},
				},
			}

			var val attr.Value
			diags := tfsdk.ValueFrom(ctx, bucketModel, resourceSchema.Type(), &val)
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
				map[string]attr.Value{"id": bucketModel.ID},
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
				DisplayName: bucketModel.Name.ValueString(),
			}) {
				return
			}
		}
	}
}
