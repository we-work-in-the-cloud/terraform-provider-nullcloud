package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

func TestInstanceAction_InvokeReportsResultingStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/instances/vsi-1/actions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var requestBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("decode action request: %v", err)
		}
		if requestBody["type"] != "stop" {
			t.Errorf("expected stop action, got %q", requestBody["type"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"vsi-1","name":"test-instance","status":"stopped"}`))
	}))
	defer server.Close()

	a := &InstanceAction{client: client.New(server.URL, "test-token")}
	var schemaResponse action.SchemaResponse
	a.Schema(context.Background(), action.SchemaRequest{}, &schemaResponse)

	configValue := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"instance_id": tftypes.String,
			"action":      tftypes.String,
		}},
		map[string]tftypes.Value{
			"instance_id": tftypes.NewValue(tftypes.String, "vsi-1"),
			"action":      tftypes.NewValue(tftypes.String, "stop"),
		},
	)
	request := action.InvokeRequest{Config: tfsdk.Config{Raw: configValue, Schema: schemaResponse.Schema}}

	var progress []action.InvokeProgressEvent
	response := &action.InvokeResponse{
		SendProgress: func(event action.InvokeProgressEvent) {
			progress = append(progress, event)
		},
	}
	a.Invoke(context.Background(), request, response)

	if response.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", response.Diagnostics)
	}

	for _, event := range progress {
		if strings.Contains(event.Message, "stopped") {
			return
		}
	}
	t.Fatalf("expected progress to report resulting status stopped; got %#v", progress)
}
