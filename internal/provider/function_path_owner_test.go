// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestFunctionPathOwner(t *testing.T) {
	// Create temporary files and directories for testing
	tempDir := t.TempDir()

	tempFile, err := os.CreateTemp(tempDir, "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }() // Ensure the file is cleaned up after the test

	tempDirNested := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(tempDirNested, 0755); err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Get the current user to validate ownership
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to retrieve current user: %v", err)
	}
	currentUsername := currentUser.Username

	// Set up the test cases
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0), // Ensure compatibility with Terraform 1.8.0+
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test case: Check owner of an existing file
				Config: `
					output "test_existing_file_owner" {
						value = provider::utilities::path_owner("` + tempFile.Name() + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_existing_file_owner", currentUsername),
				),
			},
			{
				// Test case: Check owner of an existing directory
				Config: `
					output "test_existing_dir_owner" {
						value = provider::utilities::path_owner("` + tempDirNested + `")
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test_existing_dir_owner", currentUsername),
				),
			},
			{
				// Test case: Check a non-existing path
				Config: `
					output "test_non_existing_path_owner" {
						value = provider::utilities::path_owner("/non/existing/path")
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Error: Error in function call.*Call to function "provider::utilities::path_owner" failed: Error retrieving\s+path information\.`),
			},
			{
				// Test case: Check an empty path
				Config: `
					output "test_empty_path_owner" {
						value = provider::utilities::path_owner("")
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Error: Error in function call.*Call to function "provider::utilities::path_owner" failed: Path cannot be\s+empty\.`),
			},
		},
	})
}
