package ecfollowers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type elasticFollowersProvider struct{}

type elasticFollowersProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (e elasticFollowersProvider) Metadata(
	_ context.Context,
	_ provider.MetadataRequest,
	response *provider.MetadataResponse,
) {
	response.TypeName = "ecfollowers"
}

func (e elasticFollowersProvider) Schema(
	_ context.Context,
	_ provider.SchemaRequest,
	response *provider.SchemaResponse,
) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url":      schema.StringAttribute{Required: true, Description: "Full URL of the elastic cluster."},
			"username": schema.StringAttribute{Optional: true, Description: "(Optional) Cluster username."},
			"password": schema.StringAttribute{
				Optional: true, Sensitive: true, Description: "(Optional) Cluster password.",
			},
		},
	}
}

func (e elasticFollowersProvider) Configure(
	ctx context.Context,
	request provider.ConfigureRequest,
	response *provider.ConfigureResponse,
) {
	var config elasticFollowersProviderModel
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.ResourceData = &config
}

func (e elasticFollowersProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (e elasticFollowersProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newFollowerIndexResource,
	}
}

func New() provider.Provider {
	return &elasticFollowersProvider{}
}
