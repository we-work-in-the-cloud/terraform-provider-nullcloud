package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ provider.Provider = &NullCloudProvider{}

type NullCloudProvider struct{}

func New() provider.Provider {
	return &NullCloudProvider{}
}

type providerModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

func (p *NullCloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nullcloud"
}

func (p *NullCloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "NullCloud backend URL, e.g. http://localhost:8080",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Authorization token",
			},
		},
	}
}

func (p *NullCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	c := client.New(config.URL.ValueString(), config.Token.ValueString())
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *NullCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVPCResource,
		NewSubnetResource,
		NewInstanceResource,
	}
}

func (p *NullCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
