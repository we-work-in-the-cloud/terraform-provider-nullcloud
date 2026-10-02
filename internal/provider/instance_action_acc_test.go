package provider

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

func TestAccInstanceAction_StopReportsStatusInTerraformOutput(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance test requires TF_ACC=1")
	}

	terraformPath := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if terraformPath == "" {
		var err error
		terraformPath, err = exec.LookPath("terraform")
		if err != nil {
			t.Fatalf("Terraform CLI is required (set TF_ACC_TERRAFORM_PATH or add terraform to PATH): %v", err)
		}
	}

	backend := StartBackend(t)
	defer backend.Close()

	workDir := t.TempDir()
	pluginDir := filepath.Join(workDir, "plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("create plugin directory: %v", err)
	}

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate acceptance test source")
	}
	providerRoot := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", ".."))
	providerBinary := filepath.Join(pluginDir, "terraform-provider-nullcloud")
	build := exec.Command("go", "build", "-o", providerBinary, ".")
	build.Dir = providerRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build provider binary: %v\n%s", err, output)
	}

	cliConfigPath := filepath.Join(workDir, "terraform.rc")
	cliConfig := fmt.Sprintf(`provider_installation {
  dev_overrides {
    "registry.terraform.io/we-work-in-the-cloud/nullcloud" = %q
  }
  direct {}
}
`, pluginDir)
	if err := os.WriteFile(cliConfigPath, []byte(cliConfig), 0o600); err != nil {
		t.Fatalf("write Terraform CLI config: %v", err)
	}

	config := fmt.Sprintf(`terraform {
  required_providers {
    nullcloud = {
      source = "we-work-in-the-cloud/nullcloud"
    }
  }
}

provider "nullcloud" {
  url   = %q
  token = "test-token"
}

resource "nullcloud_vpc" "test" {
  name   = "vpc-%s"
  region = "us-east"
}

resource "nullcloud_subnet" "test" {
  name       = "subnet-%s"
  vpc_id     = nullcloud_vpc.test.id
  zone       = "us-east-1"
  cidr_block = "10.0.1.0/24"
}

resource "nullcloud_instance" "test" {
  name      = "instance-%s"
  subnet_id = nullcloud_subnet.test.id
  profile   = "bx2-2x8"
  image     = "ibm-ubuntu-22-04"
}

action "nullcloud_instance_action" "stop" {
  config {
    instance_id = nullcloud_instance.test.id
    action      = "stop"
  }
}

output "instance_id" {
  value = nullcloud_instance.test.id
}
`, backend.URL(), RandomName("action"), RandomName("action"), RandomName("action"))
	configPath := filepath.Join(workDir, "main.tf")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write Terraform configuration: %v", err)
	}

	env := make([]string, 0, len(os.Environ())+1)
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "TF_CLI_CONFIG_FILE=") {
			env = append(env, value)
		}
	}
	env = append(env, "TF_CLI_CONFIG_FILE="+cliConfigPath)
	runTerraform := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(terraformPath, args...)
		cmd.Dir = workDir
		cmd.Env = env
		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output
		if err := cmd.Run(); err != nil {
			t.Fatalf("terraform %s failed: %v\n%s", strings.Join(args, " "), err, output.String())
		}
		return output.String()
	}

	runTerraform("init", "-input=false", "-no-color")
	defer func() {
		cmd := exec.Command(terraformPath, "destroy", "-auto-approve", "-input=false", "-no-color")
		cmd.Dir = workDir
		cmd.Env = env
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("terraform destroy failed: %v\n%s", err, output)
		}
	}()
	runTerraform("apply", "-auto-approve", "-input=false", "-no-color")

	instanceID := strings.TrimSpace(runTerraform("output", "-raw", "instance_id"))
	if instanceID == "" {
		t.Fatal("Terraform did not output the provisioned instance ID")
	}
	api := client.New(backend.URL(), "test-token")
	assertInstanceStatus := func(want string) {
		t.Helper()
		instance, found, err := api.GetInstance(instanceID)
		if err != nil {
			t.Fatalf("get instance %s: %v", instanceID, err)
		}
		if !found {
			t.Fatalf("instance %s not found", instanceID)
		}
		if instance.Status != want {
			t.Fatalf("expected backend instance status %q, got %q", want, instance.Status)
		}
	}
	assertInstanceStatus("running")

	applyOutput := runTerraform("apply", "-invoke=action.nullcloud_instance_action.stop", "-auto-approve", "-input=false", "-no-color")
	if !strings.Contains(applyOutput, "current status: stopped") {
		t.Fatalf("apply output did not report resulting status; output:\n%s", applyOutput)
	}
	assertInstanceStatus("stopped")
}
