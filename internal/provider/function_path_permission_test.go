package provider

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestFunctionPathPermission(t *testing.T) {
	// Create temporary files and directories for testing
	tempDir := t.TempDir()

	tempFile, err := os.CreateTemp(tempDir, "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure the file is cleaned up after the test

	// Explicitly set the file permissions to 0644
	if err := os.Chmod(tempFile.Name(), 0644); err != nil {
		t.Fatalf("Failed to set permissions on temporary file: %v", err)
	}

	tempDirNested := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(tempDirNested, 0755); err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Set up the test cases
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0), // Ensure compatibility with Terraform 1.8.0+
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test case: Check permissions of an existing file
				Config: ` 
					output "test_existing_file_permissions" {
						value = provider::utilities::path_permission("` + tempFile.Name() + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_existing_file_permissions", "0644"),
				),
			},
			{
				// Test case: Check permissions of an existing directory
				Config: ` 
					output "test_existing_dir_permissions" {
						value = provider::utilities::path_permission("` + tempDirNested + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_existing_dir_permissions", "0755"),
				),
			},
			{
				// Test case: Check a non-existing path
				Config: `
					output "test_non_existing_path_permissions" {
						value = provider::utilities::path_permission("/non/existing/path")
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Error: Error in function call.*` +
					`Call to function "provider::utilities::path_permission" failed: ` +
					`Error\s+retrieving\s+path\s+information: ` +
					`stat\s+/non/existing/path: ` +
					`no\s+such\s+file\s+or\s+directory.*`),
			},
			{
				// Test case: Check an empty path
				Config: `
					output "test_empty_path_permissions" {
						value = provider::utilities::path_permission("")
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Call to function "provider::utilities::path_permission" failed: ` +
					`Path\s+cannot\s+be\s+empty.*`),
			},
		},
	})
}
