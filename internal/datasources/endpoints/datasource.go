// Package endpoints is the `monitrs_endpoints` Terraform data source (list all endpoints).
package endpoints

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

var _ datasource.DataSource = &endpointsDataSource{}

// New returns a new instance of the monitrs_endpoints data source.
func New() datasource.DataSource { return &endpointsDataSource{} }

type endpointsDataSource struct{ client *api.Client }

// endpointItem mirrors a single endpoint returned in the list.
type endpointItem struct {
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

type endpointsModel struct {
	Limit     types.Int64    `tfsdk:"limit"`
	Endpoints []endpointItem `tfsdk:"endpoints"`
}

func (d *endpointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoints"
}

func (d *endpointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *endpointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	endpointAttrs := map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Unique identifier of the endpoint."},
		"name":             schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable name for the endpoint."},
		"url":              schema.StringAttribute{Computed: true, MarkdownDescription: "HTTPS URL that monit.rs probes."},
		"method":           schema.StringAttribute{Computed: true, MarkdownDescription: "HTTP method used for probing."},
		"interval_seconds": schema.Int64Attribute{Computed: true, MarkdownDescription: "Probe interval in seconds."},
		"is_active":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the endpoint is currently active."},
		"openapi_spec_url": schema.StringAttribute{Computed: true, MarkdownDescription: "URL of the associated OpenAPI spec, if any."},
		"heartbeat_token":  schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Opaque token for cron-liveness heartbeat pings."},
		"created_at":       schema.StringAttribute{Computed: true, MarkdownDescription: "ISO 8601 timestamp when the endpoint was created."},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns a list of all monit.rs endpoints visible to the authenticated API key.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Maximum number of endpoints to return (default 50, max 200).",
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of endpoints returned by the API.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: endpointAttrs,
				},
			},
		},
	}
}

func (d *endpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config endpointsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := int(config.Limit.ValueInt64())
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	eps, err := d.client.ListEndpoints(ctx, limit)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list endpoints", err.Error())
		return
	}

	items := make([]endpointItem, 0, len(eps))
	for _, ep := range eps {
		item := endpointItem{
			ID:              types.StringValue(ep.ID),
			Name:            types.StringValue(ep.Name),
			URL:             types.StringValue(ep.URL),
			Method:          types.StringValue(ep.Method),
			IntervalSeconds: types.Int64Value(int64(ep.IntervalSeconds)),
			IsActive:        types.BoolValue(ep.IsActive),
			CreatedAt:       types.StringValue(ep.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}
		if ep.OpenapiSpecURL != nil {
			item.OpenapiSpecURL = types.StringValue(*ep.OpenapiSpecURL)
		} else {
			item.OpenapiSpecURL = types.StringNull()
		}
		if ep.HeartbeatToken != nil {
			item.HeartbeatToken = types.StringValue(*ep.HeartbeatToken)
		} else {
			item.HeartbeatToken = types.StringNull()
		}
		items = append(items, item)
	}

	config.Limit = types.Int64Value(int64(limit))
	config.Endpoints = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
