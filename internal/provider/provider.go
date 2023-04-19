package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*externalProvider)(nil)

type externalProvider struct{}

func New() provider.Provider {
	return &externalProvider{}
}

func (p *externalProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "toolbox"
}

func (p *externalProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *externalProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *externalProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExternalResource,
	}
}

func (p *externalProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {
}
