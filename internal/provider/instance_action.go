package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

var _ action.Action = &InstanceAction{}
var _ action.ActionWithConfigure = &InstanceAction{}

type InstanceAction struct {
	client *client.Client
}

func NewInstanceAction() action.Action {
	return &InstanceAction{}
}

func (a *InstanceAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_action"
}

func (a *InstanceAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		Description: "Performs a start, stop, or restart action on a NullCloud virtual server instance.",
		Attributes: map[string]actionschema.Attribute{
			"instance_id": actionschema.StringAttribute{
				Required:    true,
				Description: "The ID of the instance to act on.",
			},
			"action": actionschema.StringAttribute{
				Required:    true,
				Description: "The action to perform: start, stop, or restart.",
			},
		},
	}
}

func (a *InstanceAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *InstanceAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config struct {
		InstanceID types.String `tfsdk:"instance_id"`
		Action     types.String `tfsdk:"action"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Performing " + config.Action.ValueString() + " on instance " + config.InstanceID.ValueString() + "...",
	})

	_, err := a.client.InstanceAction(config.InstanceID.ValueString(), config.Action.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error performing instance action", err.Error())
		return
	}
}
