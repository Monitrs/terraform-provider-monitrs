// Package status_page provides the `monitrs_status_page_endpoint` Terraform resource.
package status_page

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

var (
	_ resource.Resource                = &attachmentResource{}
	_ resource.ResourceWithImportState = &attachmentResource{}
)

// NewAttachment is the constructor registered with the provider.
func NewAttachment() resource.Resource { return &attachmentResource{} }

type attachmentResource struct{ client *api.Client }

type attachmentModel struct {
	ID           types.String `tfsdk:"id"`
	PageID       types.String `tfsdk:"page_id"`
	EndpointID   types.String `tfsdk:"endpoint_id"`
	DisplayName  types.String `tfsdk:"display_name"`
	DisplayOrder types.Int64  `tfsdk:"display_order"`
}

func (r *attachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page_endpoint"
}

func (r *attachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *attachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches an endpoint to a monit.rs status page. All fields are immutable — changing any field forces a new resource (the API has no PATCH for attachments).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server-assigned attachment ID (spe_id).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"page_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the `monitrs_status_page` to attach the endpoint to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the `monitrs_endpoint` to attach.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Label shown for this endpoint on the status page.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"display_order": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional sort order for this endpoint on the page. Server assigns a default when omitted.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace(), int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *attachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan attachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pageID := plan.PageID.ValueString()

	attachReq := struct {
		EndpointID   string `json:"endpoint_id"`
		DisplayName  string `json:"display_name"`
		DisplayOrder *int   `json:"display_order,omitempty"`
	}{
		EndpointID:  plan.EndpointID.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
	}
	if !plan.DisplayOrder.IsNull() && !plan.DisplayOrder.IsUnknown() {
		order := int(plan.DisplayOrder.ValueInt64())
		attachReq.DisplayOrder = &order
	}

	att, err := r.client.AttachEndpoint(ctx, pageID, attachReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to attach endpoint to status page", err.Error())
		return
	}

	writeAttachmentModel(&plan, att)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read does not perform a server round-trip because the API has no "get single
// attachment" endpoint. State is copied as-is.
//
// TODO: once GET /api/v1/status-pages/{id}/endpoints/{att_id} exists, do a real
// drift-detection read here. Until then, out-of-band deletes will only surface
// on the next `apply` (the server will return 404 on Delete).
func (r *attachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state attachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No-op: re-persist existing state unchanged.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op — every user-configurable field carries RequiresReplace,
// so Terraform will destroy + create rather than call Update.
func (r *attachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan attachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *attachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state attachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DetachEndpoint(ctx, state.PageID.ValueString(), state.ID.ValueString()); err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return // already gone
		}
		resp.Diagnostics.AddError("Failed to detach endpoint from status page", err.Error())
	}
}

func (r *attachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by attachment ID only. page_id must be set manually after import
	// (terraform state is incomplete until a refresh resolves page_id).
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeAttachmentModel(m *attachmentModel, att *api.StatusPageAttachment) {
	m.ID = types.StringValue(att.ID)
	m.EndpointID = types.StringValue(att.EndpointID)
	m.DisplayName = types.StringValue(att.DisplayName)
	if att.DisplayOrder != nil {
		m.DisplayOrder = types.Int64Value(int64(*att.DisplayOrder))
	} else {
		m.DisplayOrder = types.Int64Null()
	}
	// page_id is not returned by the server; preserve whatever was set in the plan.
}
