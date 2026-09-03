package status_page

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

func TestWriteStatusPageModel_populatesAllFields(t *testing.T) {
	m := &statusPageModel{}
	desc := "Live status of our public API"
	sp := &api.StatusPage{
		ID: "sp-1", Slug: "prod", Title: "Prod", Description: &desc, IsPublic: true,
	}
	writeStatusPageModel(m, sp)
	if m.ID != types.StringValue("sp-1") {
		t.Fatalf("id mismatch: %v", m.ID)
	}
	if m.Slug.ValueString() != "prod" {
		t.Fatalf("slug mismatch")
	}
	if m.Description.ValueString() != desc {
		t.Fatalf("description mismatch")
	}
	if !m.IsPublic.ValueBool() {
		t.Fatalf("is_public mismatch")
	}
}

func TestWriteStatusPageModel_nilDescription(t *testing.T) {
	m := &statusPageModel{}
	sp := &api.StatusPage{
		ID: "sp-2", Slug: "staging", Title: "Staging", Description: nil, IsPublic: false,
	}
	writeStatusPageModel(m, sp)
	if !m.Description.IsNull() {
		t.Fatalf("expected null description, got: %v", m.Description)
	}
	if m.IsPublic.ValueBool() {
		t.Fatalf("expected is_public=false")
	}
}

func TestWriteAttachmentModel_populatesAllFields(t *testing.T) {
	m := &attachmentModel{
		PageID: types.StringValue("sp-1"),
	}
	order := 3
	att := &api.StatusPageAttachment{
		ID:           "spe-42",
		EndpointID:   "ep-7",
		DisplayName:  "Auth API",
		DisplayOrder: &order,
	}
	writeAttachmentModel(m, att)
	if m.ID.ValueString() != "spe-42" {
		t.Fatalf("id mismatch: %v", m.ID)
	}
	if m.EndpointID.ValueString() != "ep-7" {
		t.Fatalf("endpoint_id mismatch")
	}
	if m.DisplayName.ValueString() != "Auth API" {
		t.Fatalf("display_name mismatch")
	}
	if m.DisplayOrder.ValueInt64() != 3 {
		t.Fatalf("display_order mismatch: %v", m.DisplayOrder)
	}
	// page_id must be preserved from the model (not overwritten by server response)
	if m.PageID.ValueString() != "sp-1" {
		t.Fatalf("page_id should be preserved, got: %v", m.PageID)
	}
}

func TestWriteAttachmentModel_nilDisplayOrder(t *testing.T) {
	m := &attachmentModel{PageID: types.StringValue("sp-1")}
	att := &api.StatusPageAttachment{
		ID:           "spe-1",
		EndpointID:   "ep-1",
		DisplayName:  "Primary",
		DisplayOrder: nil,
	}
	writeAttachmentModel(m, att)
	if !m.DisplayOrder.IsNull() {
		t.Fatalf("expected null display_order, got: %v", m.DisplayOrder)
	}
}
