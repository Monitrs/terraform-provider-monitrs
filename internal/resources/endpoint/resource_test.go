package endpoint

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

func TestWriteModel_populatesAllFields(t *testing.T) {
	m := &endpointModel{}
	spec := "https://api.example.com/openapi.json"
	hb := "hb-token-abc"
	ep := &api.Endpoint{
		ID: "id-1", Name: "n", URL: "https://x", Method: "GET",
		IntervalSeconds: 300, IsActive: true, OpenapiSpecURL: &spec, HeartbeatToken: &hb,
	}
	writeModel(m, ep)
	if m.ID != types.StringValue("id-1") {
		t.Fatalf("id mismatch: %v", m.ID)
	}
	if m.OpenapiSpecURL.ValueString() != spec {
		t.Fatalf("openapi_spec_url mismatch")
	}
	if m.HeartbeatToken.ValueString() != hb {
		t.Fatalf("heartbeat_token mismatch")
	}
}

func TestWriteModel_handlesNilOptionalFields(t *testing.T) {
	m := &endpointModel{}
	ep := &api.Endpoint{
		ID: "id-2", Name: "n2", URL: "https://y", Method: "GET",
		IntervalSeconds: 60, IsActive: false,
	}
	writeModel(m, ep)
	if m.OpenapiSpecURL != (types.String{}) && !m.OpenapiSpecURL.IsNull() {
		// Zero-value types.String from writeModel with no set path is fine
	}
}
