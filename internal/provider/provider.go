// Package provider is the Terraform provider server for monit.rs.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

// Registry: registry.terraform.io/Monitrs/monitrs
// Terraform expects the provider to satisfy provider.Provider.
var _ provider.Provider = &monitrsProvider{}

type monitrsProvider struct {
	version string
}

type providerData struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &monitrsProvider{version: version}
	}
}

func (p *monitrsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "monitrs"
	resp.Version = p.version
}

func (p *monitrsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for monit.rs (schema-aware API monitoring).",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "monit.rs API key (starts with `mrs_live_`). Falls back to the `MONITRS_API_KEY` env var.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Override the API base URL. Defaults to `https://monit.rs`. Set for self-hosted or staging deploys.",
				Optional:            true,
			},
		},
	}
}

func (p *monitrsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := data.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("MONITRS_API_KEY")
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing monit.rs API key",
			"Set the `api_key` attribute on the provider block or the `MONITRS_API_KEY` env var.",
		)
		return
	}

	opts := []api.Option{
		api.WithUserAgent("terraform-provider-monitrs/" + p.version),
	}
	if base := data.BaseURL.ValueString(); base != "" {
		opts = append(opts, api.WithBaseURL(base))
	}

	client, err := api.New(apiKey, opts...)
	if err != nil {
		resp.Diagnostics.AddError("Invalid API key", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources returns the list of resource type constructors. Wire each T2-T4 resource here.
func (p *monitrsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Filled in by T2, T3, T4.
	}
}

// DataSources returns the list of data source constructors. Wire each T5 data source here.
func (p *monitrsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Filled in by T5.
	}
}
