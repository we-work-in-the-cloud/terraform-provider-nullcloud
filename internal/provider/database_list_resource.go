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

var _ list.ListResource = &DatabaseListResource{}
var _ list.ListResourceWithConfigure = &DatabaseListResource{}

type DatabaseListResource struct {
	client *client.Client
}

type databaseListConfig struct {
	Engine types.String `tfsdk:"engine"`
}


func NewDatabaseListResource() list.ListResource {
	return &DatabaseListResource{}
}

func (r *DatabaseListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Query existing databases with optional filters.",
		Attributes: map[string]listschema.Attribute{
			"engine": listschema.StringAttribute{
				Optional:    true,
				Description: "Filter databases by engine (postgres, mysql).",
			},
		},
	}
}

func (r *DatabaseListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *DatabaseListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	if r.client == nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unconfigured Provider", "The provider hasn't been configured yet."),
		})
		return
	}

	var config databaseListConfig
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	databases, err := r.client.ListDatabases()
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to List Databases", err.Error()),
		})
		return
	}

	resp.Results = func(push func(list.ListResult) bool) {
		for _, db := range databases {
			if !config.Engine.IsNull() && db.Engine != config.Engine.ValueString() {
				continue
			}

			subnetIDValues := make([]attr.Value, len(db.SubnetIDs))
			for i, id := range db.SubnetIDs {
				subnetIDValues[i] = types.StringValue(id)
			}
			subnetIDsList, subnetDiags := types.ListValue(types.StringType, subnetIDValues)
			if subnetDiags.HasError() {
				if !push(list.ListResult{Diagnostics: subnetDiags}) {
					return
				}
				continue
			}

			dbModel := databaseModel{
				ID:        types.StringValue(db.ID),
				Name:      types.StringValue(db.Name),
				Status:    types.StringValue(db.Status),
				CRN:       types.StringValue(db.CRN),
				Engine:    types.StringValue(db.Engine),
				Version:   types.StringValue(db.Version),
				Plan:      types.StringValue(db.Plan),
				SubnetIDs: subnetIDsList,
				CreatedAt: types.StringValue(db.CreatedAt.String()),
				Endpoint:  types.StringValue(db.Endpoint),
			}

			resourceSchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id":         resourceschema.StringAttribute{Computed: true},
					"name":       resourceschema.StringAttribute{Computed: true},
					"status":     resourceschema.StringAttribute{Computed: true},
					"crn":        resourceschema.StringAttribute{Computed: true},
					"engine":     resourceschema.StringAttribute{Computed: true},
					"version":    resourceschema.StringAttribute{Computed: true},
					"plan":       resourceschema.StringAttribute{Computed: true},
					"subnet_ids": resourceschema.ListAttribute{Computed: true, ElementType: types.StringType},
					"created_at": resourceschema.StringAttribute{Computed: true},
					"endpoint":   resourceschema.StringAttribute{Computed: true},
				},
			}

			identitySchema := resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"id": resourceschema.StringAttribute{Computed: true},
				},
			}

			var val attr.Value
			diags := tfsdk.ValueFrom(ctx, dbModel, resourceSchema.Type(), &val)
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
				map[string]attr.Value{"id": dbModel.ID},
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
				DisplayName: dbModel.Name.ValueString(),
			}) {
				return
			}
		}
	}
}

// toDatabaseModel converts a client.Database to a databaseModel for Terraform.
func toDatabaseModel(db client.Database) databaseModel {
	return databaseModel{
		ID:        types.StringValue(db.ID),
		Name:      types.StringValue(db.Name),
		Engine:    types.StringValue(db.Engine),
		Status:    types.StringValue(db.Status),
		CRN:       types.StringValue(db.CRN),
		CreatedAt: types.StringValue(db.CreatedAt.String()),
	}
}

// shouldIncludeDatabase checks if a database matches the filter config.
func shouldIncludeDatabase(db client.Database, config databaseListConfig) bool {
	if !config.Engine.IsNull() && db.Engine != config.Engine.ValueString() {
		return false
	}
	return true
}
