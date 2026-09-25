package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

const testAccVpsBasic = `
data "ovh_me" "myaccount" {}

data "ovh_order_cart" "mycart" {
  ovh_subsidiary = data.ovh_me.myaccount.ovh_subsidiary
}

data "ovh_order_cart_product_plan" "vps" {
  cart_id        = data.ovh_order_cart.mycart.id
  price_capacity = "renew"
  product        = "vps"
  plan_code      = "vps-le-2-2-40"
}

resource "ovh_vps" "myvps" {
  display_name = "%s"
  netboot_mode = "rescue"

  image_id = "45b2f222-ab10-44ed-863f-720942762b6f"
  public_ssh_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDSD76EaLUzJjf70W8W2uU9FzEyl68di67Bd20qtYfBLJpFJuX/RJC9StI1y1RnXXqC1Lf/Yo+yJzvNx0iqLxCX1G7g0XYex74HkgC6a2QeNhp9M56ANZtA3TKKAbkZ1xobfhOPWpq3lEFp7dgJctcILBPL3l6OjKf6NIxHo5yF67Vy4D0nWl5utumNdWhhlX7MtVQooszLyIwPlNO+DzD3ZnJFCt2Z1jdRkhm/Oobtx17CZ+5SN23tgHXS6pLOgM6w30M11zkI510z95IAIHhRT7MbiXICkvG/0qHuSftz1j/CcHFbttNB27dH86vByumfSEgRKaoRkCqrn64IWrSsFr3Smsf7gZWLBlYLliGPyn8Tsr9bT5pRul6yTvVbfZ31RREBr1I0Lp4q++d+fIpa3LtMGRaMb9huJYy8cwW/Vfzbxsqfz9xzjIOFNcYl7J9l4cvz3hgSlai2Jgngw5ShNVlxcIKUdiynZWm09nQudlYNHgor9ID+JACzCfPkUZ8"

  ovh_subsidiary = data.ovh_order_cart.mycart.ovh_subsidiary
  plan = [
    {
      duration     = "P1M"
      plan_code    = data.ovh_order_cart_product_plan.vps.plan_code
      pricing_mode = "default"

      configuration = [
        {
          label = "vps_datacenter"
          value = "WAW"
        },
        {
          label = "vps_os"
          value = "Debian 10"
        }
      ]
    }
  ]
}
`

const testAccVpsDoNotSendPassword = `
data "ovh_me" "myaccount" {}

data "ovh_order_cart" "mycart" {
  ovh_subsidiary = data.ovh_me.myaccount.ovh_subsidiary
}

data "ovh_order_cart_product_plan" "vps" {
  cart_id        = data.ovh_order_cart.mycart.id
  price_capacity = "renew"
  product        = "vps"
  plan_code      = "vps-le-2-2-40"
}

resource "ovh_vps" "myvps" {
  display_name = "%s"
  netboot_mode = "rescue"

  image_id = "45b2f222-ab10-44ed-863f-720942762b6f"
  public_ssh_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDSD76EaLUzJjf70W8W2uU9FzEyl68di67Bd20qtYfBLJpFJuX/RJC9StI1y1RnXXqC1Lf/Yo+yJzvNx0iqLxCX1G7g0XYex74HkgC6a2QeNhp9M56ANZtA3TKKAbkZ1xobfhOPWpq3lEFp7dgJctcILBPL3l6OjKf6NIxHo5yF67Vy4D0nWl5utumNdWhhlX7MtVQooszLyIwPlNO+DzD3ZnJFCt2Z1jdRkhm/Oobtx17CZ+5SN23tgHXS6pLOgM6w30M11zkI510z95IAIHhRT7MbiXICkvG/0qHuSftz1j/CcHFbttNB27dH86vByumfSEgRKaoRkCqrn64IWrSsFr3Smsf7gZWLBlYLliGPyn8Tsr9bT5pRul6yTvVbfZ31RREBr1I0Lp4q++d+fIpa3LtMGRaMb9huJYy8cwW/Vfzbxsqfz9xzjIOFNcYl7J9l4cvz3hgSlai2Jgngw5ShNVlxcIKUdiynZWm09nQudlYNHgor9ID+JACzCfPkUZ8"
  do_not_send_password = true

  ovh_subsidiary = data.ovh_order_cart.mycart.ovh_subsidiary
  plan = [
    {
      duration     = "P1M"
      plan_code    = data.ovh_order_cart_product_plan.vps.plan_code
      pricing_mode = "default"

      configuration = [
        {
          label = "vps_datacenter"
          value = "WAW"
        },
        {
          label = "vps_os"
          value = "Debian 10"
        }
      ]
    }
  ]
}
`

const testAccVpsReinstallImageOnly = `
data "ovh_me" "myaccount" {}

data "ovh_order_cart" "mycart" {
  ovh_subsidiary = data.ovh_me.myaccount.ovh_subsidiary
}

data "ovh_order_cart_product_plan" "vps" {
  cart_id        = data.ovh_order_cart.mycart.id
  price_capacity = "renew"
  product        = "vps"
  plan_code      = "vps-le-2-2-40"
}

resource "ovh_vps" "myvps" {
  display_name = "%s"
  netboot_mode = "rescue"

  image_id = "45b2f222-ab10-44ed-863f-720942762b6f"

  ovh_subsidiary = data.ovh_order_cart.mycart.ovh_subsidiary
  plan = [
    {
      duration     = "P1M"
      plan_code    = data.ovh_order_cart_product_plan.vps.plan_code
      pricing_mode = "default"

      configuration = [
        {
          label = "vps_datacenter"
          value = "WAW"
        },
        {
          label = "vps_os"
          value = "Debian 10"
        }
      ]
    }
  ]
}
`

// Update rebuilds its state from GET /vps/{serviceName}, which never returns
// install options, and merges the plan into it. An install option the merge
// drops comes back null against a known planned value, which Terraform
// rejects as an inconsistent result after apply.
func TestVpsModelMergeWithKeepsPlannedInstallOptions(t *testing.T) {
	fromAPI := VpsModel{
		ServiceName:       ovhtypes.NewTfStringValue("vps-test.vps.ovh.net"),
		ImageId:           ovhtypes.NewTfStringNull(),
		PublicSSHKey:      ovhtypes.NewTfStringNull(),
		DoNotSendPassword: ovhtypes.TfBoolValue{BoolValue: basetypes.NewBoolNull()},
		PostInstallScript: ovhtypes.NewTfStringNull(),
	}
	planned := VpsModel{
		ImageId:           ovhtypes.NewTfStringValue("45b2f222-ab10-44ed-863f-720942762b6f"),
		PublicSSHKey:      ovhtypes.NewTfStringNull(),
		DoNotSendPassword: ovhtypes.TfBoolValue{BoolValue: basetypes.NewBoolValue(true)},
		PostInstallScript: ovhtypes.NewTfStringValue("#!/bin/bash\necho first boot\n"),
	}

	fromAPI.MergeWith(&planned)

	if !fromAPI.DoNotSendPassword.ValueBool() {
		t.Fatalf("do_not_send_password = %v, want the planned true", fromAPI.DoNotSendPassword)
	}
	if got := fromAPI.PostInstallScript.ValueString(); got != planned.PostInstallScript.ValueString() {
		t.Fatalf("post_install_script = %q, want the planned script", got)
	}
	if got := fromAPI.ImageId.ValueString(); got != planned.ImageId.ValueString() {
		t.Fatalf("image_id = %q, want the planned image", got)
	}
}

func TestVpsPostInstallScriptIsASensitiveInstallOption(t *testing.T) {
	attribute := VpsResourceSchema(context.Background()).Attributes["post_install_script"]
	if !attribute.IsSensitive() {
		t.Fatal("post_install_script must be sensitive: first-boot scripts carry credentials")
	}

	scriptOnly := VpsModel{
		ImageId:           ovhtypes.NewTfStringNull(),
		PublicSSHKey:      ovhtypes.NewTfStringNull(),
		PostInstallScript: ovhtypes.NewTfStringValue("#!/bin/bash\n"),
	}
	if !installOptionsHasBeenSet(scriptOnly) {
		t.Fatal("a post_install_script alone is an install option")
	}
	if summary, _ := validateInstallOptions(scriptOnly, VpsModel{}); summary == "" {
		t.Fatal("a post_install_script without image_id must be rejected")
	}

	previous := VpsModel{
		ImageId:           ovhtypes.NewTfStringValue("45b2f222-ab10-44ed-863f-720942762b6f"),
		PublicSSHKey:      ovhtypes.NewTfStringNull(),
		PostInstallScript: ovhtypes.NewTfStringValue("#!/bin/bash\n"),
	}
	cleared := previous
	cleared.PostInstallScript = ovhtypes.NewTfStringNull()
	if _, details := validateInstallOptions(cleared, previous); details != fmt.Sprintf("You cannot set to null a previously non-null value (%s)", "post_install_script") {
		t.Fatalf("clearing post_install_script: %q", details)
	}
}

func TestVpsToInstallOptionsSendsPostInstallScript(t *testing.T) {
	model := VpsModel{
		ImageId:           ovhtypes.NewTfStringValue("45b2f222-ab10-44ed-863f-720942762b6f"),
		PublicSSHKey:      ovhtypes.NewTfStringNull(),
		DoNotSendPassword: ovhtypes.TfBoolValue{BoolValue: basetypes.NewBoolValue(true)},
		PostInstallScript: ovhtypes.NewTfStringValue("#!/bin/bash\necho first boot\n"),
	}

	body, err := json.Marshal(model.ToInstallOptions())
	if err != nil {
		t.Fatal(err)
	}

	var sent map[string]any
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatal(err)
	}
	if sent["postInstallScript"] != "#!/bin/bash\necho first boot\n" || sent["doNotSendPassword"] != true || sent["imageId"] != "45b2f222-ab10-44ed-863f-720942762b6f" {
		t.Fatalf("rebuild body = %s", body)
	}
	if !installOptionsHasChanged(model, VpsModel{ImageId: model.ImageId, PublicSSHKey: model.PublicSSHKey, PostInstallScript: ovhtypes.NewTfStringValue("#!/bin/bash\n")}) {
		t.Fatal("a changed post_install_script must reinstall the VPS")
	}
}

// Adopts an existing VPS and reinstalls it with a first-boot script and no
// password e-mail. THIS WIPES THE VPS, so it runs only against the one named
// in OVH_VPS_REINSTALL, never the shared OVH_VPS fixture. The last step drops
// it from state without destroying it: the resource's Delete terminates the
// service. Needs Terraform >= 1.7 (removed blocks).
func TestAccResourceVps_importReinstallPostInstallScript(t *testing.T) {
	serviceName := os.Getenv("OVH_VPS_REINSTALL")
	imageID := os.Getenv("OVH_VPS_IMAGE_ID")
	vps := fmt.Sprintf(`
resource "ovh_vps" "myvps" {
  plan = []

  image_id             = "%s"
  do_not_send_password = true
  post_install_script  = <<-EOT
    #!/bin/bash
    echo terraform-provider-ovh > /root/post-install-ran
  EOT

  lifecycle {
    prevent_destroy = true
  }
}
`, imageID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckCredentials(t)
			checkEnvOrSkip(t, "OVH_VPS_REINSTALL")
			checkEnvOrSkip(t, "OVH_VPS_IMAGE_ID")
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("import {\n  to = ovh_vps.myvps\n  id = %q\n}\n%s", serviceName, vps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ovh_vps.myvps", "service_name", serviceName),
					resource.TestCheckResourceAttr("ovh_vps.myvps", "image_id", imageID),
					resource.TestCheckResourceAttr("ovh_vps.myvps", "do_not_send_password", "true"),
					resource.TestCheckResourceAttrSet("ovh_vps.myvps", "post_install_script"),
					resource.TestCheckResourceAttr("ovh_vps.myvps", "state", "running"),
				),
			},
			{
				Config:   vps,
				PlanOnly: true,
			},
			{
				Config: `
removed {
  from = ovh_vps.myvps
  lifecycle {
    destroy = false
  }
}
`,
			},
		},
	})
}

func TestAccResourceVps_basic(t *testing.T) {
	displayName := acctest.RandomWithPrefix(test_prefix)
	config := fmt.Sprintf(
		testAccVpsBasic,
		displayName,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckOrderVPS(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "netboot_mode", "rescue"),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "display_name", displayName),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "public_ssh_key", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDSD76EaLUzJjf70W8W2uU9FzEyl68di67Bd20qtYfBLJpFJuX/RJC9StI1y1RnXXqC1Lf/Yo+yJzvNx0iqLxCX1G7g0XYex74HkgC6a2QeNhp9M56ANZtA3TKKAbkZ1xobfhOPWpq3lEFp7dgJctcILBPL3l6OjKf6NIxHo5yF67Vy4D0nWl5utumNdWhhlX7MtVQooszLyIwPlNO+DzD3ZnJFCt2Z1jdRkhm/Oobtx17CZ+5SN23tgHXS6pLOgM6w30M11zkI510z95IAIHhRT7MbiXICkvG/0qHuSftz1j/CcHFbttNB27dH86vByumfSEgRKaoRkCqrn64IWrSsFr3Smsf7gZWLBlYLliGPyn8Tsr9bT5pRul6yTvVbfZ31RREBr1I0Lp4q++d+fIpa3LtMGRaMb9huJYy8cwW/Vfzbxsqfz9xzjIOFNcYl7J9l4cvz3hgSlai2Jgngw5ShNVlxcIKUdiynZWm09nQudlYNHgor9ID+JACzCfPkUZ8"),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "image_id", os.Getenv("OVH_VPS_IMAGE_ID")),
				),
			},
		},
	})
}

func TestAccResourceVps_doNotSendPassword(t *testing.T) {
	displayName := acctest.RandomWithPrefix(test_prefix)
	config := fmt.Sprintf(
		testAccVpsBasic,
		displayName,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckOrderVPS(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "netboot_mode", "rescue"),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "display_name", displayName),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "public_ssh_key", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDSD76EaLUzJjf70W8W2uU9FzEyl68di67Bd20qtYfBLJpFJuX/RJC9StI1y1RnXXqC1Lf/Yo+yJzvNx0iqLxCX1G7g0XYex74HkgC6a2QeNhp9M56ANZtA3TKKAbkZ1xobfhOPWpq3lEFp7dgJctcILBPL3l6OjKf6NIxHo5yF67Vy4D0nWl5utumNdWhhlX7MtVQooszLyIwPlNO+DzD3ZnJFCt2Z1jdRkhm/Oobtx17CZ+5SN23tgHXS6pLOgM6w30M11zkI510z95IAIHhRT7MbiXICkvG/0qHuSftz1j/CcHFbttNB27dH86vByumfSEgRKaoRkCqrn64IWrSsFr3Smsf7gZWLBlYLliGPyn8Tsr9bT5pRul6yTvVbfZ31RREBr1I0Lp4q++d+fIpa3LtMGRaMb9huJYy8cwW/Vfzbxsqfz9xzjIOFNcYl7J9l4cvz3hgSlai2Jgngw5ShNVlxcIKUdiynZWm09nQudlYNHgor9ID+JACzCfPkUZ8"),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "do_not_send_password", "true"),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "image_id", os.Getenv("OVH_VPS_IMAGE_ID")),
				),
			},
		},
	})
}

func TestAccResourceVps_reinstallImageOnly(t *testing.T) {
	displayName := acctest.RandomWithPrefix(test_prefix)
	config := fmt.Sprintf(
		testAccVpsReinstallImageOnly,
		displayName,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckOrderVPS(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "netboot_mode", "rescue"),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "display_name", displayName),
					resource.TestCheckResourceAttr(
						"ovh_vps.myvps", "image_id", "45b2f222-ab10-44ed-863f-720942762b6f"),
					resource.TestCheckResourceAttrSet("ovh_vps.myvps", "id"),
				),
			},
		},
	})
}
