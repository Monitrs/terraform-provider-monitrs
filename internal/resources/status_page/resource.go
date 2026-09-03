// Package status_page provides the `monitrs_status_page` Terraform resource.
package status_page

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

var (
	_ resource.Resource                = &statusPageResource{}
	_ resource.ResourceWithImportState = &statusPageResource{}
)

// NewStatusPage is the constructor registered with the provider.
func NewStatusPage() resource.Resource { return &statusPageResource{} }

type statusPageResource struct{ client *api.Client }

type statusPageModel struct {
	ID          types.String `tfsdk:"id"`
	Slug        types.String `tfsdk:"slug"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	IsPublic    types.Bool   `tfsdk:"is_public"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *statusPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *statusPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A monit.rs status page. Attach endpoints via `monitrs_status_page_endpoint`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL slug for the status page (e.g. `my-api`). Immutable — changing it forces a new resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable title displayed on the status page.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description shown below the title.",
			},
			"is_public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the status page is publicly accessible (default: true).",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *statusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan statusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := api.StatusPageCreate{
		Slug:         plan.Slug.ValueString(),
		Title:        plan.Title.ValueString(),
		IsPublic:     plan.IsPublic.ValueBool(),
		EndpointIDs:  []string{},
		DisplayNames: []string{},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		s := plan.Description.ValueString()
		body.Description = &s
	}

	sp, err := r.client.CreateStatusPage(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create status page", err.Error())
		return
	}

	writeStatusPageModel(&plan, sp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state statusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp, err := r.client.GetStatusPage(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read status page", err.Error())
		return
	}
	writeStatusPageModel(&state, sp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan statusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build a partial-update map — only send fields that are set.
	// Skip slug: it's RequiresReplace so it will never appear in an update.
	body := map[string]any{}
	if !plan.Title.IsNull() && !plan.Title.IsUnknown() {
		body["title"] = plan.Title.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		body["is_public"] = plan.IsPublic.ValueBool()
	}

	sp, err := r.client.UpdateStatusPage(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update status page", err.Error())
		return
	}
	writeStatusPageModel(&plan, sp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state statusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteStatusPage(ctx, state.ID.ValueString()); err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return // already gone
		}
		resp.Diagnostics.AddError("Failed to delete status page", err.Error())
	}
}

func (r *statusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeStatusPageModel(m *statusPageModel, sp *api.StatusPage) {
	m.ID = types.StringValue(sp.ID)
	m.Slug = types.StringValue(sp.Slug)
	m.Title = types.StringValue(sp.Title)
	m.IsPublic = types.BoolValue(sp.IsPublic)
	m.CreatedAt = types.StringValue(sp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if sp.Description != nil {
		m.Description = types.StringValue(*sp.Description)
	} else {
		m.Description = types.StringNull()
	}
}
