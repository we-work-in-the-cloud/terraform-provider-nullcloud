package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ datasource.DataSource = &RegionsDataSource{}

var zoneAttrTypes = map[string]attr.Type{
	"name": types.StringType,
}

var regionAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"zones": types.ListType{ElemType: types.ObjectType{AttrTypes: zoneAttrTypes}},
}

type RegionsDataSource struct {
	client *client.Client
}

func NewRegionsDataSource() datasource.DataSource {
	return &RegionsDataSource{}
}

type regionsDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Regions types.List   `tfsdk:"regions"`
}

func (d *RegionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *RegionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the list of valid NullCloud regions and their availability zones.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"regions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of available regions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Region name (e.g. us-east).",
						},
						"zones": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Availability zones within this region.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:    true,
										Description: "Zone name (e.g. us-east-1).",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *RegionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *RegionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	regions, err := d.client.ListRegions()
	if err != nil {
		resp.Diagnostics.AddError("Error listing regions", err.Error())
		return
	}

	regionObjs := make([]attr.Value, len(regions))
	for i, r := range regions {
		zoneObjs := make([]attr.Value, len(r.Zones))
		for j, z := range r.Zones {
			obj, diags := types.ObjectValue(zoneAttrTypes, map[string]attr.Value{
				"name": types.StringValue(z.Name),
			})
			resp.Diagnostics.Append(diags...)
			zoneObjs[j] = obj
		}
		zonesList, diags := types.ListValue(types.ObjectType{AttrTypes: zoneAttrTypes}, zoneObjs)
		resp.Diagnostics.Append(diags...)

		regionObj, diags := types.ObjectValue(regionAttrTypes, map[string]attr.Value{
			"name":  types.StringValue(r.Name),
			"zones": zonesList,
		})
		resp.Diagnostics.Append(diags...)
		regionObjs[i] = regionObj
	}
	if resp.Diagnostics.HasError() {
		return
	}

	regionsList, diags := types.ListValue(types.ObjectType{AttrTypes: regionAttrTypes}, regionObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := regionsDataSourceModel{
		ID:      types.StringValue("regions"),
		Regions: regionsList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
