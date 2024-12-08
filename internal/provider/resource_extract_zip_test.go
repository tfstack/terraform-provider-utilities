package provider

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/pkg/errors"
)

// Function to download a file from URL.
func downloadFile(url, dest string) error {
	// Ensure the destination directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}

	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to get URL")
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write data to file")
	}

	return nil
}

func TestResourceUtilitiesExtractZip(t *testing.T) {
	// Define file paths as variables
	zipDownloadDir := fmt.Sprintf("/tmp/download-dir-%s", uuid.NewString())
	extractedDir := fmt.Sprintf("/tmp/extracted-dir-%s", uuid.NewString())
	backupDir := fmt.Sprintf("/tmp/backup-dir-%s", uuid.NewString())
	zipFilePath := zipDownloadDir + "/jq_1.7.zip" // Full path to the ZIP file
	zipURL := "https://github.com/platformfuzz/rpm-builder/archive/refs/tags/jq_1.7.zip"

	// Ensure cleanup even on test failure
	defer func() {
		os.RemoveAll(zipDownloadDir)
		os.RemoveAll(extractedDir)
		os.RemoveAll(backupDir)
	}()

	// Download the zip file before starting the test
	t.Log("Downloading ZIP file for extraction...")
	err := downloadFile(zipURL, zipFilePath)
	if err != nil {
		t.Fatalf("Failed to download ZIP file: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0), // Skip for versions below 1.8.0
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					terraform {
						required_providers {
							utilities = {
								source = "hashicorp.com/tfstack/utilities"
							}
						}
					}

					provider "utilities" {}

					resource "utilities_extract_zip" "example1" {
						destination = "%s"
						url         = "%s"
					}
					`, extractedDir, zipURL),
				Check: resource.TestCheckFunc(func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["utilities_extract_zip.example1"]
					if !ok {
						return fmt.Errorf("resource not found: utilities_extract_zip.example1")
					}

					if rs.Primary.Attributes["destination"] != extractedDir {
						return fmt.Errorf("expected destination to be '%s', got '%s'", extractedDir, rs.Primary.Attributes["destination"])
					}

					if rs.Primary.Attributes["url"] != zipURL {
						return fmt.Errorf("expected url to be '%s', got '%s'", zipURL, rs.Primary.Attributes["url"])
					}

					urlRegex := regexp.MustCompile(`^https?://.+\.zip$`)
					if !urlRegex.MatchString(rs.Primary.Attributes["url"]) {
						return fmt.Errorf("url '%s' does not match the expected pattern", rs.Primary.Attributes["url"])
					}

					if rs.Primary.Attributes["source"] != "" {
						return fmt.Errorf("expected 'source' to be empty, but found '%s'", rs.Primary.Attributes["source"])
					}

					return nil
				}),
			},
			{
				Config: fmt.Sprintf(`
					terraform {
						required_providers {
							utilities = {
								source = "hashicorp.com/tfstack/utilities"
							}
						}
					}

					provider "utilities" {}

					resource "utilities_extract_zip" "example2" {
						destination = "%s"
						source      = "%s"
					}
					`, backupDir, zipFilePath),
				Check: resource.TestCheckFunc(func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["utilities_extract_zip.example2"]
					if !ok {
						return fmt.Errorf("resource not found: utilities_extract_zip.example2")
					}

					if rs.Primary.Attributes["destination"] != backupDir {
						return fmt.Errorf("expected destination to be '%s', but found '%s'", backupDir, rs.Primary.Attributes["destination"])
					}

					if rs.Primary.Attributes["url"] != "" {
						return fmt.Errorf("expected url to be empty, but found '%s'", rs.Primary.Attributes["url"])
					}

					if rs.Primary.Attributes["source"] != zipFilePath {
						return fmt.Errorf("expected source to be '%s', got '%s'", zipFilePath, rs.Primary.Attributes["source"])
					}

					return nil
				}),
			},
		},
	})
}
