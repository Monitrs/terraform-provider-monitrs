// Package endpoint is the `monitrs_endpoint` Terraform resource.
package endpoint

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

var (
	_ resource.Resource                = &endpointResource{}
	_ resource.ResourceWithImportState = &endpointResource{}
)

func New() resource.Resource { return &endpointResource{} }

type endpointResource struct{ client *api.Client }

type endpointModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	URL             types.String `tfsdk:"url"`
	Method          types.String `tfsdk:"method"`
	IntervalSeconds types.Int64  `tfsdk:"interval_seconds"`
	OpenapiSpecURL  types.String `tfsdk:"openapi_spec_url"`
	IsActive        types.Bool   `tfsdk:"is_active"`
	HeartbeatToken  types.String `tfsdk:"heartbeat_token"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (r *endpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r *endpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "expected *api.Client")
		return
	}
	r.client = c
}

func (r *endpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A monitored endpoint. monit.rs probes this on the configured interval.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":             schema.StringAttribute{Required: true, MarkdownDescription: "Human-readable name for the endpoint."},
			"url":              schema.StringAttribute{Required: true, MarkdownDescription: "HTTPS URL to probe."},
			"method":           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "HTTP method (default: `GET`)."},
			"interval_seconds": schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Probe interval in seconds. Minimum depends on tier (Free 900, Developer 300, Starter 60, Pro/Scale 30)."},
			"openapi_spec_url": schema.StringAttribute{Optional: true, MarkdownDescription: "URL of the OpenAPI spec for this endpoint. Enables extra contract checks."},
			"is_active":        schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether the endpoint is active (default: true)."},
			"heartbeat_token":  schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Opaque token for cron-liveness heartbeat pings. Use in `POST https://monit.rs/api/v1/heartbeats/{token}`."},
			"created_at":       schema.StringAttribute{Computed: true},
		},
	}
}

func (r *endpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan endpointModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := api.EndpointCreate{
		Name:            plan.Name.ValueString(),
		URL:             plan.URL.ValueString(),
		Method:          plan.Method.ValueString(),
		IntervalSeconds: int(plan.IntervalSeconds.ValueInt64()),
	}
	if !plan.OpenapiSpecURL.IsNull() && !plan.OpenapiSpecURL.IsUnknown() {
		s := plan.OpenapiSpecURL.ValueString()
		body.OpenapiSpecURL = &s
	}

	ep, err := r.client.CreateEndpoint(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create endpoint", err.Error())
		return
	}

	writeModel(&plan, ep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *endpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state endpointModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := r.client.GetEndpoint(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read endpoint", err.Error())
		return
	}
	writeModel(&state, ep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *endpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan endpointModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := api.EndpointUpdate{}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		s := plan.Name.ValueString()
		body.Name = &s
	}
	if !plan.URL.IsNull() && !plan.URL.IsUnknown() {
		s := plan.URL.ValueString()
		body.URL = &s
	}
	if !plan.Method.IsNull() && !plan.Method.IsUnknown() {
		s := plan.Method.ValueString()
		body.Method = &s
	}
	if !plan.IntervalSeconds.IsNull() && !plan.IntervalSeconds.IsUnknown() {
		i := int(plan.IntervalSeconds.ValueInt64())
		body.IntervalSeconds = &i
	}
	if !plan.IsActive.IsNull() && !plan.IsActive.IsUnknown() {
		b := plan.IsActive.ValueBool()
		body.IsActive = &b
	}

	ep, err := r.client.UpdateEndpoint(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update endpoint", err.Error())
		return
	}
	writeModel(&plan, ep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *endpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state endpointModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEndpoint(ctx, state.ID.ValueString()); err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return // already gone
		}
		resp.Diagnostics.AddError("Failed to delete endpoint", err.Error())
	}
}

func (r *endpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeModel(m *endpointModel, ep *api.Endpoint) {
	m.ID = types.StringValue(ep.ID)
	m.Name = types.StringValue(ep.Name)
	m.URL = types.StringValue(ep.URL)
	m.Method = types.StringValue(ep.Method)
	m.IntervalSeconds = types.Int64Value(int64(ep.IntervalSeconds))
	m.IsActive = types.BoolValue(ep.IsActive)
	m.CreatedAt = types.StringValue(ep.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if ep.HeartbeatToken != nil {
		m.HeartbeatToken = types.StringValue(*ep.HeartbeatToken)
	}
	if ep.OpenapiSpecURL != nil {
		m.OpenapiSpecURL = types.StringValue(*ep.OpenapiSpecURL)
	}
}
