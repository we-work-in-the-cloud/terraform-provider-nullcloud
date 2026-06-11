package provider

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// BackendManager starts and stops a backend process for system tests.
type BackendManager struct {
	cmd        *exec.Cmd
	url        string
	port       string
	binaryPath string
	t          *testing.T
}

// StartBackend starts a backend instance on a random available port.
// The BACKEND_BINARY environment variable must be set to the path of the backend binary.
// The caller must call Close() to clean up, typically in a defer statement.
// It returns the backend URL (e.g., "http://localhost:12345").
func StartBackend(t *testing.T) *BackendManager {
	t.Helper()

	// Get backend binary path from environment
	binaryPath := os.Getenv("BACKEND_BINARY")
	if binaryPath == "" {
		t.Fatalf("BACKEND_BINARY environment variable not set")
	}

	if info, err := os.Stat(binaryPath); err != nil || info.IsDir() {
		t.Fatalf("BACKEND_BINARY not found or not a file: %s", binaryPath)
	}

	// Find a random available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	portStr := fmt.Sprintf("%d", port)
	url := fmt.Sprintf("http://localhost:%s", portStr)

	// Start the backend process
	cmd := exec.Command(binaryPath, "-p", portStr)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start backend: %v", err)
	}

	bm := &BackendManager{
		cmd:        cmd,
		url:        url,
		port:       portStr,
		binaryPath: binaryPath,
		t:          t,
	}

	// Wait for backend to be ready
	if !bm.waitReady() {
		t.Fatalf("backend at %s failed to start", url)
	}

	return bm
}

// waitReady polls the backend until it responds or times out.
func (bm *BackendManager) waitReady() bool {
	deadline := time.Now().Add(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", bm.port))
			if err == nil {
				resp.Close()
				return true
			}
			if time.Now().After(deadline) {
				return false
			}
		}
	}
}

// URL returns the backend URL (e.g., "http://localhost:12345").
func (bm *BackendManager) URL() string {
	return bm.url
}

// Port returns the backend port number as a string.
func (bm *BackendManager) Port() string {
	return bm.port
}

// Close stops the backend process. It's safe to call multiple times.
func (bm *BackendManager) Close() error {
	if bm.cmd == nil || bm.cmd.Process == nil {
		return nil
	}

	// Try graceful shutdown first
	_ = bm.cmd.Process.Signal(os.Interrupt)

	// Wait a bit for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- bm.cmd.Wait()
	}()

	select {
	case <-time.After(2 * time.Second):
		// Force kill if it doesn't shutdown gracefully
		_ = bm.cmd.Process.Kill()
	case <-done:
		return nil
	}

	return bm.cmd.Wait()
}

// GetProviderConfig returns a provider configuration block using the backend manager.
func (bm *BackendManager) GetProviderConfig() string {
	return fmt.Sprintf(`
provider "nullcloud" {
  url   = "%s"
  token = "test-token"
}
`, bm.URL())
}

// ProtoV6ProviderFactories returns the provider factories for acceptance tests.
func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"nullcloud": providerserver.NewProtocol6WithError(New()),
	}
}

// RandomName generates a random resource name for testing.
func RandomName(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
}
