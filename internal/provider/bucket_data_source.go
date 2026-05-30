package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &BucketDataSource{}

type BucketDataSource struct {
	client *client.Client
}

func NewBucketDataSource() datasource.DataSource {
	return &BucketDataSource{}
}

func (d *BucketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (d *BucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a NullCloud object storage bucket by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the bucket.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"region": schema.StringAttribute{
				Computed: true,
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
		},
	}
}

func (d *BucketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *BucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, found, err := d.client.GetBucket(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Bucket not found", fmt.Sprintf("Bucket with ID %q not found", data.ID.ValueString()))
		return
	}

	data.Name = types.StringValue(b.Name)
	data.Region = types.StringValue(b.Region)
	data.Status = types.StringValue(b.Status)
	data.CRN = types.StringValue(b.CRN)
	data.CreatedAt = types.StringValue(b.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
