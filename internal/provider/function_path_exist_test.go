package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestPathExists(t *testing.T) {
	// Create temporary files and directories for testing
	tempDir := t.TempDir()

	tempFile, err := os.CreateTemp(tempDir, "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure the file is cleaned up after the test

	tempDirNested := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(tempDirNested, 0755); err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	tempSymlink := filepath.Join(tempDir, "testsymlink")
	if err := os.Symlink(tempFile.Name(), tempSymlink); err != nil {
		t.Fatalf("Failed to create temporary symlink: %v", err)
	}

	brokenSymlink := filepath.Join(tempDir, "brokensymlink")
	if err := os.Symlink("nonexistent", brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Set up the test cases
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0), // Ensure compatibility with Terraform 1.8.0+
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test case: Check an existing file
				Config: `
					output "test_existing_file" {
						value = provider::utilities::path_exists("` + tempFile.Name() + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_existing_file", "true"),
				),
			},
			{
				// Test case: Check an existing directory
				Config: `
					output "test_existing_dir" {
						value = provider::utilities::path_exists("` + tempDirNested + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_existing_dir", "true"),
				),
			},
			{
				// Test case: Check a non-existing path
				Config: `
					output "test_non_existing_path" {
						value = provider::utilities::path_exists("/non/existing/path")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_non_existing_path", "false"),
				),
			},
			{
				// Test case: Check a valid symlink
				Config: `
					output "test_valid_symlink" {
						value = provider::utilities::path_exists("` + tempSymlink + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_valid_symlink", "true"),
				),
			},
			{
				// Test case: Check a broken symlink
				Config: `
					output "test_broken_symlink" {
						value = provider::utilities::path_exists("` + brokenSymlink + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_broken_symlink", "false"),
				),
			},
			{
				// Test case: Check an empty path
				Config: `
					output "test_empty_path" {
						value = provider::utilities::path_exists("")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_empty_path", "false"),
				),
			},
		},
	})
}
