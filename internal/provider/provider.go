// Copyright (c) Plex, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure PwPusherProvider satisfies various provider interfaces.
var _ provider.Provider = &PwPusherProvider{}
var _ provider.ProviderWithFunctions = &PwPusherProvider{}

// PwPusherProvider defines the provider implementation.
type PwPusherProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PwPusherProviderModel describes the provider data model.
type PwPusherProviderModel struct {
	Url types.String `tfsdk:"url"`
}

type ProviderData struct {
	client *http.Client
	url    types.String
}

func (p *PwPusherProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pwpusher"
	resp.Version = p.version
}

func (p *PwPusherProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL for the pwpusher service",
				Optional:            true,
			},
		},
	}
}

func (p *PwPusherProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PwPusherProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	if data.Url.IsNull() {
		data.Url = types.StringValue("https://pwpush.com")
	}

	providerData := ProviderData{
		client: http.DefaultClient,
		url:    data.Url,
	}
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *PwPusherProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTextResource,
	}
}

func (p *PwPusherProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *PwPusherProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PwPusherProvider{
			version: version,
		}
	}
}
