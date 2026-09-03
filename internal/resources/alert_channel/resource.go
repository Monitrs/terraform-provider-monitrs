// Package alert_channel provides the `monitrs_alert_channel` Terraform resource.
package alert_channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

var (
	_ resource.Resource                = &alertChannelResource{}
	_ resource.ResourceWithImportState = &alertChannelResource{}
)

// New is the constructor registered with the provider.
func New() resource.Resource { return &alertChannelResource{} }

type alertChannelResource struct{ client *api.Client }

type alertChannelModel struct {
	ID          types.String `tfsdk:"id"`
	ChannelType types.String `tfsdk:"channel_type"`
	Config      types.Map    `tfsdk:"config"`
	MinSeverity types.String `tfsdk:"min_severity"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *alertChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_channel"
}

func (r *alertChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *alertChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a monit.rs alert channel. All user-configurable fields are immutable — changing any field forces a new resource (the API has no PATCH for alert channels).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server-assigned alert channel ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"channel_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Delivery mechanism for alerts. Valid values: `email`, `slack`, `discord`, `teams`, `pagerduty`, `webhook`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"config": schema.MapAttribute{
				Required:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
				MarkdownDescription: "Channel-specific configuration as a map of string key-value pairs (e.g. `url` for Slack/Discord/webhook channels). In v0.1 only string values are supported — non-string API fields are not representable in this schema.",
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"min_severity": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Minimum alert severity to deliver to this channel (e.g. `warning`, `critical`). When omitted, the server default applies.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO 8601 timestamp when the alert channel was created.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *alertChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert types.Map to map[string]any (values are strings from the schema).
	config := map[string]any{}
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		var m map[string]string
		resp.Diagnostics.Append(plan.Config.ElementsAs(ctx, &m, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range m {
			config[k] = v
		}
	}

	body := api.AlertChannelCreate{
		ChannelType: plan.ChannelType.ValueString(),
		Config:      config,
	}
	if !plan.MinSeverity.IsNull() && !plan.MinSeverity.IsUnknown() {
		s := plan.MinSeverity.ValueString()
		body.MinSeverity = &s
	}

	ch, err := r.client.CreateAlertChannel(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create alert channel", err.Error())
		return
	}

	// Write server response back to state.
	writeModel(ctx, &plan, ch, resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channels, err := r.client.ListAlertChannels(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read alert channels", err.Error())
		return
	}
	var found *api.AlertChannel
	for i := range channels {
		if channels[i].ID == state.ID.ValueString() {
			found = &channels[i]
			break
		}
	}
	if found == nil {
		// Channel deleted out-of-band; drop from state.
		resp.State.RemoveResource(ctx)
		return
	}
	writeModel(ctx, &state, found, resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op — every user-configurable field carries RequiresReplace,
// so Terraform will destroy + create rather than call Update.
func (r *alertChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAlertChannel(ctx, state.ID.ValueString()); err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return // already gone
		}
		resp.Diagnostics.AddError("Failed to delete alert channel", err.Error())
	}
}

func (r *alertChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// writeModel populates m from the API response ch.
// ch.Config map[string]any values are converted to strings; non-string values
// are stringified with fmt.Sprintf so the types.Map (element type: String)
// constraint is satisfied.
func writeModel(ctx context.Context, m *alertChannelModel, ch *api.AlertChannel, diags diag.Diagnostics) {
	m.ID = types.StringValue(ch.ID)
	m.ChannelType = types.StringValue(ch.ChannelType)
	m.CreatedAt = types.StringValue(ch.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if ch.MinSeverity != nil {
		m.MinSeverity = types.StringValue(*ch.MinSeverity)
	} else {
		m.MinSeverity = types.StringNull()
	}

	// Convert map[string]any → map[string]string for the types.Map.
	strMap := make(map[string]string, len(ch.Config))
	for k, v := range ch.Config {
		if s, ok := v.(string); ok {
			strMap[k] = s
		} else {
			strMap[k] = fmt.Sprintf("%v", v)
		}
	}
	mapVal, d := types.MapValueFrom(ctx, types.StringType, strMap)
	diags.Append(d...)
	m.Config = mapVal
}
