// Package endpoint is the `monitrs_endpoint` Terraform data source (lookup by id).
package endpoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

var _ datasource.DataSource = &endpointDataSource{}

// New returns a new instance of the monitrs_endpoint data source.
func New() datasource.DataSource { return &endpointDataSource{} }

type endpointDataSource struct{ client *api.Client }

type endpointModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	URL             types.String `tfsdk:"url"`
	Method          types.String `tfsdk:"method"`
	IntervalSeconds types.Int64  `tfsdk:"interval_seconds"`
	IsActive        types.Bool   `tfsdk:"is_active"`
	OpenapiSpecURL  types.String `tfsdk:"openapi_spec_url"`
	HeartbeatToken  types.String `tfsdk:"heartbeat_token"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (d *endpointDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (d *endpointDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "expected *api.Client")
		return
	}
	d.client = c
}

func (d *endpointDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single monit.rs endpoint by ID.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Required: true, MarkdownDescription: "Unique identifier of the endpoint."},
			"name":             schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable name for the endpoint."},
			"url":              schema.StringAttribute{Computed: true, MarkdownDescription: "HTTPS URL that monit.rs probes."},
			"method":           schema.StringAttribute{Computed: true, MarkdownDescription: "HTTP method used for probing."},
			"interval_seconds": schema.Int64Attribute{Computed: true, MarkdownDescription: "Probe interval in seconds."},
			"is_active":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the endpoint is currently active."},
			"openapi_spec_url": schema.StringAttribute{Computed: true, MarkdownDescription: "URL of the associated OpenAPI spec, if any."},
			"heartbeat_token":  schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Opaque token for cron-liveness heartbeat pings."},
			"created_at":       schema.StringAttribute{Computed: true, MarkdownDescription: "ISO 8601 timestamp when the endpoint was created."},
		},
	}
}

func (d *endpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config endpointModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := d.client.GetEndpoint(ctx, config.ID.ValueString())
	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.Diagnostics.AddError("Endpoint not found", fmt.Sprintf("no endpoint with id %s", config.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Failed to read endpoint", err.Error())
		return
	}

	config.ID = types.StringValue(ep.ID)
	config.Name = types.StringValue(ep.Name)
	config.URL = types.StringValue(ep.URL)
	config.Method = types.StringValue(ep.Method)
	config.IntervalSeconds = types.Int64Value(int64(ep.IntervalSeconds))
	config.IsActive = types.BoolValue(ep.IsActive)
	config.CreatedAt = types.StringValue(ep.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if ep.OpenapiSpecURL != nil {
		config.OpenapiSpecURL = types.StringValue(*ep.OpenapiSpecURL)
	} else {
		config.OpenapiSpecURL = types.StringNull()
	}

	if ep.HeartbeatToken != nil {
		config.HeartbeatToken = types.StringValue(*ep.HeartbeatToken)
	} else {
		config.HeartbeatToken = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
