package alert_channel

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

func TestWriteModel_populatesAllFields(t *testing.T) {
	m := &alertChannelModel{}
	severity := "critical"
	ch := &api.AlertChannel{
		ID:          "ch-1",
		ChannelType: "slack",
		Config:      map[string]any{"url": "https://hooks.slack.com/x"},
		MinSeverity: &severity,
	}
	var d diag.Diagnostics
	writeModel(context.Background(), m, ch, d)
	if m.ID != types.StringValue("ch-1") {
		t.Fatalf("id mismatch: %v", m.ID)
	}
	if m.ChannelType.ValueString() != "slack" {
		t.Fatalf("channel_type mismatch")
	}
	if m.MinSeverity.ValueString() != "critical" {
		t.Fatalf("min_severity mismatch")
	}
}

func TestWriteModel_nilMinSeverity(t *testing.T) {
	m := &alertChannelModel{}
	ch := &api.AlertChannel{
		ID:          "ch-2",
		ChannelType: "webhook",
		Config:      map[string]any{"url": "https://example.com/hook"},
		MinSeverity: nil,
	}
	var d diag.Diagnostics
	writeModel(context.Background(), m, ch, d)
	if !m.MinSeverity.IsNull() {
		t.Fatalf("expected null min_severity, got %v", m.MinSeverity)
	}
}

func TestWriteModel_nonStringConfigValue(t *testing.T) {
	m := &alertChannelModel{}
	ch := &api.AlertChannel{
		ID:          "ch-3",
		ChannelType: "pagerduty",
		Config:      map[string]any{"retry_count": 3},
		MinSeverity: nil,
	}
	var d diag.Diagnostics
	writeModel(context.Background(), m, ch, d)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics errors: %v", d)
	}
	elems := m.Config.Elements()
	v, ok := elems["retry_count"]
	if !ok {
		t.Fatal("expected retry_count key in config map")
	}
	if v.(types.String).ValueString() != "3" {
		t.Fatalf("expected stringified int '3', got %v", v)
	}
}
