package provider

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccExamples validates all Terraform examples in the examples directory.
// Dynamically discovers and tests all example directories (excluding data-sources and actions).
func TestAccExamples(t *testing.T) {
	bm := StartBackend(t)
	defer bm.Close()

	examplesRoot := "../../examples"
	exampleDirs := findExampleDirs(t, examplesRoot)

	t.Logf("Starting validation of %d Terraform examples", len(exampleDirs))

	for _, examplePath := range exampleDirs {
		absExamplesRoot, _ := filepath.Abs(examplesRoot)
		rel, _ := filepath.Rel(absExamplesRoot, examplePath)
		exampleName := filepath.Base(examplePath)

		t.Run(exampleName, func(t *testing.T) {
			t.Logf("Validating example: %s (%s)", exampleName, rel)
			config, err := loadExampleConfig(t, examplePath, bm)
			if err != nil {
				t.Fatalf("failed to load example: %v", err)
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
				Steps: []resource.TestStep{
					{
						Config: config,
					},
				},
			})
			t.Logf("✓ %s validated successfully", exampleName)
		})
	}

	t.Logf("✓ All %d examples passed validation", len(exampleDirs))
}

// findExampleDirs recursively finds all directories containing .tf files.
// Skips data-sources and actions which are documentation examples, not executable.
func findExampleDirs(t *testing.T, root string) []string {
	t.Helper()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	var dirs []string
	filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip data-sources and actions (they're documentation, not executable examples)
		if d.IsDir() && (strings.Contains(path, "data-sources") || strings.Contains(path, "actions")) {
			return fs.SkipDir
		}

		if !d.IsDir() {
			return nil
		}

		// Check if this directory has .tf files
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		hasTfFiles := false
		hasSubdirWithTf := false

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
				hasTfFiles = true
			} else if entry.IsDir() {
				subEntries, err := os.ReadDir(filepath.Join(path, entry.Name()))
				if err == nil {
					for _, subEntry := range subEntries {
						if !subEntry.IsDir() && strings.HasSuffix(subEntry.Name(), ".tf") {
							hasSubdirWithTf = true
							break
						}
					}
				}
			}
		}

		// Add if it's a leaf directory (has .tf files but subdirs don't)
		if hasTfFiles && !hasSubdirWithTf {
			dirs = append(dirs, path)
		}

		return nil
	})

	if len(dirs) == 0 {
		t.Fatalf("no example directories found in %s", absRoot)
	}

	return dirs
}

// loadExampleConfig reads all .tf files from an example directory and merges them,
// replacing provider and terraform blocks with one configured for the test backend.
func loadExampleConfig(t *testing.T, exampleDir string, bm *BackendManager) (string, error) {
	t.Helper()

	entries, err := os.ReadDir(exampleDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var tfFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}

		tfFiles = append(tfFiles, filepath.Join(exampleDir, entry.Name()))
	}

	if len(tfFiles) == 0 {
		return "", fmt.Errorf("no .tf files found")
	}

	configParts := []string{bm.GetProviderConfig()}

	for _, tfFile := range tfFiles {
		content, err := os.ReadFile(tfFile)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		filtered := stripBlockLines(string(content), []string{"terraform", "provider"})

		if strings.TrimSpace(filtered) != "" {
			configParts = append(configParts, filtered)
		}
	}

	return strings.Join(configParts, "\n\n"), nil
}

// stripBlockLines removes entire blocks matching the given block types from HCL.
func stripBlockLines(config string, blockTypes []string) string {
	lines := strings.Split(config, "\n")
	var result []string
	inBlock := false
	blockDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if starting a block we want to skip
		if !inBlock {
			for _, blockType := range blockTypes {
				if strings.HasPrefix(trimmed, blockType) && strings.Contains(trimmed, "{") {
					inBlock = true
					break
				}
			}
			if inBlock {
				blockDepth = 0
				continue
			}
		}

		if inBlock {
			blockDepth += strings.Count(line, "{")
			blockDepth -= strings.Count(line, "}")

			if blockDepth < 0 {
				inBlock = false
				blockDepth = 0
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
