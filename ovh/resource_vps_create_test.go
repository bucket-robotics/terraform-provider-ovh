package ovh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/ovhwrap"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
	"go.uber.org/ratelimit"
)

// An order API that refuses the cart, as one without the right to order
// does, and records every request it gets.
type refusingOrderAPI struct {
	mu       sync.Mutex
	requests []string
}

func (f *refusingOrderAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, r.Method+" "+r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	if strings.HasSuffix(r.URL.Path, "/auth/time") {
		_ = json.NewEncoder(w).Encode(time.Now().Unix())
		return
	}
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "This call has not been granted"})
}

func createVps(t *testing.T, api http.Handler, set map[string]any) resource.CreateResponse {
	t.Helper()
	ctx := context.Background()
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	client, err := ovh.NewClient(server.URL, "a", "b", "c")
	if err != nil {
		t.Fatal(err)
	}
	r := &vpsResource{config: &Config{OVHClient: ovhwrap.NewClient(client, ratelimit.NewUnlimited())}}

	schema := VpsResourceSchema(ctx)
	null := tftypes.NewValue(schema.Type().TerraformType(ctx), nil)
	plan := tfsdk.Plan{Schema: schema, Raw: null}
	for attribute, value := range set {
		if diags := plan.SetAttribute(ctx, path.Root(attribute), value); diags.HasError() {
			t.Fatal(diags)
		}
	}
	planValue := NewPlanValueMust(PlanValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"configuration": ovhtypes.TfListNestedValue[PlanConfigurationValue]{ListValue: basetypes.NewListValueMust(PlanConfigurationValue{}.Type(ctx), []attr.Value{})},
		"duration":      ovhtypes.NewTfStringValue("P1M"),
		"item_id":       ovhtypes.TfInt64Value{Int64Value: basetypes.NewInt64Null()},
		"plan_code":     ovhtypes.NewTfStringValue("vps-2027-model2"),
		"pricing_mode":  ovhtypes.NewTfStringValue("default"),
		"quantity":      ovhtypes.TfInt64Value{Int64Value: basetypes.NewInt64Null()},
	})
	plans := ovhtypes.TfListNestedValue[PlanValue]{ListValue: basetypes.NewListValueMust(PlanValue{}.Type(ctx), []attr.Value{planValue})}
	if diags := plan.SetAttribute(ctx, path.Root("plan"), plans); diags.HasError() {
		t.Fatal(diags)
	}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schema, Raw: null}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
	return resp
}

// A refused order saves nothing: a VPS saved with no service name fails
// every later Read, and one never ordered must not be in state.
func TestVpsCreateSavesNothingWhenTheOrderIsRefused(t *testing.T) {
	api := &refusingOrderAPI{}
	resp := createVps(t, api, map[string]any{"display_name": "node-4", "ovh_subsidiary": "US"})
	if !resp.Diagnostics.HasError() {
		t.Fatal("Create succeeded, want the refused order to fail it")
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("Create saved state after a refused order: %v", resp.State.Raw)
	}
	if len(api.requests) == 0 || !strings.Contains(strings.Join(api.requests, " "), "POST /1.0/order/cart") && !strings.Contains(strings.Join(api.requests, " "), "POST /order/cart") {
		t.Fatalf("Create never asked for a cart (%v): %v", api.requests, resp.Diagnostics)
	}
	for _, request := range api.requests {
		if strings.Contains(request, "/pay") || strings.Contains(request, "/checkout") {
			t.Fatalf("a refused cart still reached %s", request)
		}
	}
}

// A VPS being ordered takes no install options: a rebuild failing inside
// Create would leave a paid VPS tainted, which only a terminate replaces.
func TestVpsCreateRefusesInstallOptionsBeforeOrdering(t *testing.T) {
	api := &refusingOrderAPI{}
	resp := createVps(t, api, map[string]any{"display_name": "node-4", "ovh_subsidiary": "US", "image_id": "image-1"})
	if !resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
		t.Fatalf("Create with install options: errors %v, state %v", resp.Diagnostics, resp.State.Raw)
	}
	for _, request := range api.requests {
		if strings.Contains(request, "/order/") {
			t.Fatalf("Create ordered before refusing its install options: %s", request)
		}
	}
}
