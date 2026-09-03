// Package me is the `monitrs_me` Terraform data source (authenticated user info).
package me

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Monitrs/terraform-provider-monitrs/internal/api"
)

var _ datasource.DataSource = &meDataSource{}

// New returns a new instance of the monitrs_me data source.
func New() datasource.DataSource { return &meDataSource{} }

type meDataSource struct{ client *api.Client }

type meModel struct {
	Email    types.String `tfsdk:"email"`
	Tier     types.String `tfsdk:"tier"`
	AuthKind types.String `tfsdk:"auth_kind"`
	Scopes   types.List   `tfsdk:"scopes"`
}

func (d *meDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_me"
}

func (d *meDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *meDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns information about the authenticated monit.rs account (whoami).",
		Attributes: map[string]schema.Attribute{
			"email":     schema.StringAttribute{Computed: true, MarkdownDescription: "Email address of the authenticated account."},
			"tier":      schema.StringAttribute{Computed: true, MarkdownDescription: "Subscription tier of the account (e.g. `free`, `developer`, `pro`)."},
			"auth_kind": schema.StringAttribute{Computed: true, MarkdownDescription: "Authentication kind used (e.g. `api_key`, `session`)."},
			"scopes": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of permission scopes granted to the API key.",
			},
		},
	}
}

func (d *meDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config meModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	me, err := d.client.Me(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read account info", err.Error())
		return
	}

	config.Email = types.StringValue(me.Email)
	config.Tier = types.StringValue(me.Tier)
	config.AuthKind = types.StringValue(me.AuthKind)

	scopeVals := make([]attr.Value, len(me.Scopes))
	for i, s := range me.Scopes {
		scopeVals[i] = types.StringValue(s)
	}
	scopesList, diags := types.ListValue(types.StringType, scopeVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Scopes = scopesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
