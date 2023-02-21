package ecfollowers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &followerIndexResource{}
	_ resource.ResourceWithConfigure = &followerIndexResource{}
)

type followerIndexResource struct {
	providerConfig *elasticFollowersProviderModel
}

type followerIndexResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Index         types.String `tfsdk:"index"`
	RemoteCluster types.String `tfsdk:"remote_cluster"`
	LeaderIndex   types.String `tfsdk:"leader_index"`
	CleanUpIndex  types.Bool   `tfsdk:"cleanup_index"`
	LastUpdated   types.String `tfsdk:"last_updated"`
}

type elasticFollowerIndexCreate struct {
	RemoteCluster string `json:"remote_cluster"`
	LeaderIndex   string `json:"leader_index"`
}

type elasticFollowerIndexInfo struct {
	FollowerIndices []elasticFollowerIndice `json:"follower_indices"`
}

type elasticFollowerIndice struct {
	FollowIndex   string `json:"follow_index"`
	RemoteCluster string `json:"remote_cluster"`
	LeaderIndex   string `json:"leader_index"`
}

func (f *followerIndexResource) Configure(
	_ context.Context,
	request resource.ConfigureRequest,
	response *resource.ConfigureResponse,
) {
	if request.ProviderData == nil {
		return
	}

	cfg, ok := request.ProviderData.(*elasticFollowersProviderModel)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *elasticFollowersProviderModel, got: %T. Please report this issue to the provider developers.",
				request.ProviderData,
			),
		)
	}
	f.providerConfig = cfg
}

func newFollowerIndexResource() resource.Resource {
	return &followerIndexResource{}
}

func (f *followerIndexResource) Metadata(
	_ context.Context,
	request resource.MetadataRequest,
	response *resource.MetadataResponse,
) {
	response.TypeName = request.ProviderTypeName + "_follower_index"
}

func (f *followerIndexResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	response *resource.SchemaResponse,
) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"index": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "Elastic Index to create as a follower.",
			},
			"remote_cluster": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "Name of the remote cluster.",
			},
			"leader_index": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "Name of the index to follow on the remote cluster.",
			},
			"cleanup_index": schema.BoolAttribute{
				Optional:    true,
				Description: "If set to true, it will also delete the local index when removing the follower. (Defaults to true)",
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (f *followerIndexResource) Create(
	ctx context.Context,
	request resource.CreateRequest,
	response *resource.CreateResponse,
) {
	var plan followerIndexResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data, err := json.Marshal(
		elasticFollowerIndexCreate{
			RemoteCluster: plan.RemoteCluster.ValueString(),
			LeaderIndex:   plan.LeaderIndex.ValueString(),
		},
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to build payload to create follower index",
			"Could not create payload to create follower index, unexpected error: "+err.Error(),
		)
		return
	}

	esClient := f.getClient()

	followResp, err := sendElasticRequest(
		ctx,
		esClient,
		&esapi.CCRFollowRequest{Index: plan.Index.ValueString(), Body: bytes.NewReader(data)},
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to create follower index",
			"Could not create follower index, unexpected error trying to call "+f.providerConfig.URL.String()+" :"+err.Error(),
		)
		return
	}
	followResp.Body.Close()

	plan.ID = types.StringValue(
		plan.RemoteCluster.ValueString() + "_" + plan.LeaderIndex.ValueString() + "_" + plan.Index.ValueString(),
	)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (f *followerIndexResource) Read(
	ctx context.Context,
	request resource.ReadRequest,
	response *resource.ReadResponse,
) {
	var state followerIndexResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	esClient := f.getClient()

	esResponse, err := sendElasticRequest(
		ctx,
		esClient,
		esapi.CCRFollowInfoRequest{Index: []string{state.Index.ValueString()}},
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to fetch follower index info",
			"Could not fetch follower index info, unexpected error: "+err.Error(),
		)
		return
	}
	defer esResponse.Body.Close()

	var info elasticFollowerIndexInfo
	if err := json.NewDecoder(esResponse.Body).Decode(&info); err != nil {
		response.Diagnostics.AddError(
			"Failed to parse follower index info",
			"Could not parse follower index info, unexpected error: "+err.Error(),
		)
		return
	}

	if len(info.FollowerIndices) == 0 {
		response.Diagnostics.AddError(
			"Failed to fetch follower index info",
			"Not data was returned from elasticsearch",
		)
		return
	}

	state.LeaderIndex = types.StringValue(info.FollowerIndices[0].LeaderIndex)
	state.RemoteCluster = types.StringValue(info.FollowerIndices[0].RemoteCluster)

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (f *followerIndexResource) Update(
	ctx context.Context,
	request resource.UpdateRequest,
	response *resource.UpdateResponse,
) {
	var plan followerIndexResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (f *followerIndexResource) Delete(
	ctx context.Context,
	request resource.DeleteRequest,
	response *resource.DeleteResponse,
) {
	var state followerIndexResourceModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	esClient := f.getClient()

	esCloseResponse, err := sendElasticRequest(
		ctx,
		esClient,
		esapi.IndicesCloseRequest{Index: []string{state.Index.ValueString()}},
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to delete follower index",
			"Could not delete follower index, unexpected error: "+err.Error(),
		)
		return
	}
	esCloseResponse.Body.Close()

	esPauseResp, err := sendElasticRequest(ctx, esClient, esapi.CCRPauseFollowRequest{Index: state.Index.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to pause follower index",
			"Could not pause follower index, unexpected error: "+err.Error(),
		)
		return
	}
	esPauseResp.Body.Close()

	esUnfflowResp, err := sendElasticRequest(ctx, esClient, esapi.CCRUnfollowRequest{Index: state.Index.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to delete follower index",
			"Could not delete follower index, unexpected error: "+err.Error(),
		)
		return
	}
	esUnfflowResp.Body.Close()

	cleanUp := true
	if !state.CleanUpIndex.IsNull() {
		cleanUp = state.CleanUpIndex.ValueBool()
	}

	if cleanUp {
		esDeleteIndexResp, err := sendElasticRequest(
			ctx,
			esClient,
			esapi.IndicesDeleteRequest{Index: []string{state.Index.ValueString()}},
		)
		if err != nil {
			response.Diagnostics.AddError(
				"Failed to delete index",
				"Could not delete index, unexpected error: "+err.Error(),
			)
			return
		}
		esDeleteIndexResp.Body.Close()
	}
}

func (f *followerIndexResource) getClient() *elasticsearch.TypedClient {
	esConfig := elasticsearch.Config{Addresses: []string{f.providerConfig.URL.ValueString()}}
	if !f.providerConfig.Username.IsNull() && !f.providerConfig.Password.IsNull() {
		esConfig.Username = f.providerConfig.Username.ValueString()
		esConfig.Password = f.providerConfig.Password.ValueString()
	}
	esClient, _ := elasticsearch.NewTypedClient(esConfig)

	return esClient
}

func sendElasticRequest(ctx context.Context, esClient *elasticsearch.TypedClient, req esapi.Request) (
	*esapi.Response, error,
) {
	resp, err := req.Do(ctx, esClient)
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		esErr := fmt.Errorf("failed with error: %s", resp.String())
		resp.Body.Close()

		return nil, esErr
	}

	return resp, nil

}
