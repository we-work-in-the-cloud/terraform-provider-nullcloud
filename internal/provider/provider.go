package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ provider.Provider = &NullCloudProvider{}
var _ provider.ProviderWithActions = &NullCloudProvider{}

type NullCloudProvider struct{}

func New() provider.Provider {
	return &NullCloudProvider{}
}

type providerModel struct {
	URL    types.String `tfsdk:"url"`
	Token  types.String `tfsdk:"token"`
	Region types.String `tfsdk:"region"`
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
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Default region for all resources. Defaults to us-east if not set.",
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
	if r := config.Region.ValueString(); r != "" {
		c.DefaultRegion = r
	}
	resp.DataSourceData = c
	resp.ResourceData = c
	resp.ActionData = c
}

func (p *NullCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVPCResource,
		NewSubnetResource,
		NewInstanceResource,
		NewLoadBalancerResource,
		NewBucketResource,
		NewDatabaseResource,
		NewKubernetesClusterResource,
	}
}

func (p *NullCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVPCDataSource,
		NewSubnetDataSource,
		NewInstanceDataSource,
		NewLoadBalancerDataSource,
		NewBucketDataSource,
		NewDatabaseDataSource,
		NewKubernetesClusterDataSource,
		NewRegionsDataSource,
	}
}

func (p *NullCloudProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		NewInstanceAction,
	}
}
