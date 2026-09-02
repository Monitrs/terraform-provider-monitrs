package api

import "time"

// Endpoint mirrors the /api/v1/endpoints response shape (extra server-added
// fields are ignored — the client only reads what it declares).
type Endpoint struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	Method          string    `json:"method"`
	IntervalSeconds int       `json:"interval_seconds"`
	IsActive        bool      `json:"is_active"`
	OpenapiSpecURL  *string   `json:"openapi_spec_url,omitempty"`
	HeartbeatToken  *string   `json:"heartbeat_token,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type EndpointCreate struct {
	Name            string  `json:"name"`
	URL             string  `json:"url"`
	Method          string  `json:"method,omitempty"`
	IntervalSeconds int     `json:"interval_seconds,omitempty"`
	OpenapiSpecURL  *string `json:"openapi_spec_url,omitempty"`
}

type EndpointUpdate struct {
	Name            *string `json:"name,omitempty"`
	URL             *string `json:"url,omitempty"`
	Method          *string `json:"method,omitempty"`
	IntervalSeconds *int    `json:"interval_seconds,omitempty"`
	OpenapiSpecURL  *string `json:"openapi_spec_url,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
}

type StatusPage struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
}

type StatusPageCreate struct {
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Description  *string  `json:"description,omitempty"`
	IsPublic     bool     `json:"is_public"`
	EndpointIDs  []string `json:"endpoint_ids"`
	DisplayNames []string `json:"display_names"`
}

type StatusPageAttachment struct {
	ID           string `json:"id"`
	EndpointID   string `json:"endpoint_id"`
	DisplayName  string `json:"display_name"`
	DisplayOrder *int   `json:"display_order,omitempty"`
}

type AlertChannel struct {
	ID          string         `json:"id"`
	ChannelType string         `json:"channel_type"`
	Config      map[string]any `json:"config"`
	MinSeverity *string        `json:"min_severity,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type AlertChannelCreate struct {
	ChannelType string         `json:"channel_type"`
	Config      map[string]any `json:"config"`
	MinSeverity *string        `json:"min_severity,omitempty"`
}

type Me struct {
	Email    string   `json:"email"`
	Tier     string   `json:"tier"`
	Scopes   []string `json:"scopes"`
	AuthKind string   `json:"auth_kind"`
}
