package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &DatabaseDataSource{}

type DatabaseDataSource struct {
	client *client.Client
}

func NewDatabaseDataSource() datasource.DataSource {
	return &DatabaseDataSource{}
}

func (d *DatabaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *DatabaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a NullCloud managed database by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the database.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"engine": schema.StringAttribute{
				Computed: true,
			},
			"version": schema.StringAttribute{
				Computed: true,
			},
			"plan": schema.StringAttribute{
				Computed: true,
			},
			"subnet_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"crn": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Connection endpoint for the database (e.g., db-xxxxx.db.nullcloud.internal:5432).",
			},
		},
	}
}

func (d *DatabaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data databaseModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, found, err := d.client.GetDatabase(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading database", err.Error())
		return
	}
	if !found {
		reportNotFoundDataSource(resp, "Database", data.ID.ValueString())
		return
	}

	data.Name = types.StringValue(db.Name)
	data.Engine = types.StringValue(db.Engine)
	data.Version = types.StringValue(db.Version)
	data.Plan = types.StringValue(db.Plan)
	data.Status = types.StringValue(db.Status)
	data.CRN = types.StringValue(db.CRN)
	data.SubnetIDs = listOfStrings(db.SubnetIDs)
	data.CreatedAt = types.StringValue(db.CreatedAt.String())
	data.Endpoint = types.StringValue(db.Endpoint)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
