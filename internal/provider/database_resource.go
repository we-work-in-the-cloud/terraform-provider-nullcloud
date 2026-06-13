package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ resource.Resource = &DatabaseResource{}

type DatabaseResource struct {
	client *client.Client
}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type databaseModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	CRN       types.String `tfsdk:"crn"`
	Engine    types.String `tfsdk:"engine"`
	Version   types.String `tfsdk:"version"`
	Plan      types.String `tfsdk:"plan"`
	SubnetIDs types.List   `tfsdk:"subnet_ids"`
	CreatedAt types.String `tfsdk:"created_at"`
	Endpoint  types.String `tfsdk:"endpoint"`
}

func (r *DatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NullCloud managed database.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the database instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"engine": schema.StringAttribute{
				Required:    true,
				Description: "Database engine. Must be postgres, mysql, or mariadb.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "Version of the database engine (e.g. 15, 8.0).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan": schema.StringAttribute{
				Required:    true,
				Description: "Instance plan size. Must be small, medium, or large.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "List of subnet IDs where the database nodes will be deployed.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
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
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Connection endpoint for the database (e.g., db-xxxxx.db.nullcloud.internal:5432).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data databaseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var subnetIDs []string
	resp.Diagnostics.Append(data.SubnetIDs.ElementsAs(ctx, &subnetIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := r.client.CreateDatabase(
		data.Name.ValueString(),
		data.Engine.ValueString(),
		data.Version.ValueString(),
		data.Plan.ValueString(),
		subnetIDs,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", err.Error())
		return
	}

	data.ID = types.StringValue(db.ID)
	data.Status = types.StringValue(db.Status)
	data.CRN = types.StringValue(db.CRN)
	data.SubnetIDs = stringsToList(db.SubnetIDs)
	data.CreatedAt = types.StringValue(db.CreatedAt.String())
	data.Endpoint = types.StringValue(db.Endpoint)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data databaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, found, err := r.client.GetDatabase(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading database", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(db.Name)
	data.Engine = types.StringValue(db.Engine)
	data.Version = types.StringValue(db.Version)
	data.Plan = types.StringValue(db.Plan)
	data.Status = types.StringValue(db.Status)
	data.CRN = types.StringValue(db.CRN)
	data.SubnetIDs = stringsToList(db.SubnetIDs)
	data.Endpoint = types.StringValue(db.Endpoint)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All mutable attributes require replace; Update is never called.
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data databaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDatabase(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting database", err.Error())
	}
}
