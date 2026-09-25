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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/ovhwrap"
	"go.uber.org/ratelimit"
)

// A fake /vps API whose GET fails once right after a rebuild is accepted,
// the way a reinstall that outlives the wait fails the apply.
type fakeVpsAPI struct {
	mu            sync.Mutex
	rebuilds      int
	failAfterPost bool
}

func (f *fakeVpsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	reply := func(code int, body any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(body)
	}
	switch {
	case strings.HasSuffix(r.URL.Path, "/auth/time"):
		reply(http.StatusOK, time.Now().Unix())
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/rebuild"):
		f.rebuilds++
		f.failAfterPost = true
		reply(http.StatusOK, map[string]any{"id": 1, "state": "todo"})
	case r.Method == http.MethodPut:
		reply(http.StatusOK, nil)
	case r.Method == http.MethodGet && f.failAfterPost:
		f.failAfterPost = false
		reply(http.StatusServiceUnavailable, map[string]string{"message": "Service Unavailable"})
	default:
		reply(http.StatusOK, map[string]any{"serviceName": "vps-test", "displayName": "vps-test", "state": "running"})
	}
}

func TestVpsUpdateRecordsAcceptedInstallOptionsWhenTheWaitFails(t *testing.T) {
	ctx := context.Background()
	api := &fakeVpsAPI{}
	server := httptest.NewServer(api)
	defer server.Close()

	client, err := ovh.NewClient(server.URL, "a", "b", "c")
	if err != nil {
		t.Fatal(err)
	}
	r := &vpsResource{config: &Config{OVHClient: ovhwrap.NewClient(client, ratelimit.NewUnlimited())}}

	schema := VpsResourceSchema(ctx)
	null := tftypes.NewValue(schema.Type().TerraformType(ctx), nil)
	prior := tfsdk.State{Schema: schema, Raw: null}
	plan := tfsdk.Plan{Schema: schema, Raw: null}
	for attr, value := range map[string]any{"service_name": "vps-test", "display_name": "vps-test"} {
		if diags := prior.SetAttribute(ctx, path.Root(attr), value); diags.HasError() {
			t.Fatal(diags)
		}
	}
	for attr, value := range map[string]any{
		"service_name":         "vps-test",
		"display_name":         "node-1",
		"image_id":             "image-1",
		"post_install_script":  "#!/bin/bash\ntrue\n",
		"do_not_send_password": true,
	} {
		if diags := plan.SetAttribute(ctx, path.Root(attr), value); diags.HasError() {
			t.Fatal(diags)
		}
	}

	// The framework hands Update the prior state as its response state.
	resp := resource.UpdateResponse{State: tfsdk.State{Schema: schema, Raw: prior.Raw.Copy()}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Update succeeded, want the failed wait to fail the apply")
	}
	if api.rebuilds != 1 {
		t.Fatalf("rebuild POSTs = %d, want 1", api.rebuilds)
	}
	var saved VpsModel
	if diags := resp.State.Get(ctx, &saved); diags.HasError() {
		t.Fatal(diags)
	}
	if saved.ImageId.ValueString() != "image-1" || saved.PostInstallScript.ValueString() != "#!/bin/bash\ntrue\n" || !saved.DoNotSendPassword.ValueBool() {
		t.Fatalf("state after the failed wait: image_id = %s, post_install_script set = %t, do_not_send_password = %s; want the accepted install options, so the next apply does not reinstall again",
			saved.ImageId, !saved.PostInstallScript.IsNull(), saved.DoNotSendPassword)
	}
	var planned VpsModel
	if diags := plan.Get(ctx, &planned); diags.HasError() {
		t.Fatal(diags)
	}
	if installOptionsHasChanged(planned, saved) {
		t.Fatal("the same plan against the saved state would reinstall the VPS again")
	}
}
