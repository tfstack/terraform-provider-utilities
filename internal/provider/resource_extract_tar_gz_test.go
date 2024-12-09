package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestResourceUtilitiesExtractTarGz(t *testing.T) {
	// Define file paths as variables
	zipDownloadDir := fmt.Sprintf("/tmp/download-dir-%s", uuid.NewString())
	extractedDir := fmt.Sprintf("/tmp/extracted-dir-%s", uuid.NewString())
	backupDir := fmt.Sprintf("/tmp/backup-dir-%s", uuid.NewString())
	tagGzFilePath := zipDownloadDir + "/jq_1.7.tar.gz" // Full path to the TarGz file
	tarGzURL := "https://github.com/platformfuzz/rpm-builder/archive/refs/tags/jq_1.7.tar.gz"

	// Ensure cleanup even on test failure
	defer func() {
		os.RemoveAll(zipDownloadDir)
		os.RemoveAll(extractedDir)
		os.RemoveAll(backupDir)
	}()

	// Download the TarGz file before starting the test
	t.Log("Downloading TarGz file for extraction...")
	err := downloadFile(tarGzURL, tagGzFilePath)
	if err != nil {
		t.Fatalf("Failed to download TarGz file: %v", err)
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

					resource "utilities_extract_tar_gz" "example1" {
						destination = "%s"
						url         = "%s"
					}
					`, extractedDir, tarGzURL),
				Check: resource.TestCheckFunc(func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["utilities_extract_tar_gz.example1"]
					if !ok {
						return fmt.Errorf("resource not found: utilities_extract_tar_gz.example1")
					}

					if rs.Primary.Attributes["destination"] != extractedDir {
						return fmt.Errorf("expected destination to be '%s', got '%s'", extractedDir, rs.Primary.Attributes["destination"])
					}

					if rs.Primary.Attributes["url"] != tarGzURL {
						return fmt.Errorf("expected url to be '%s', got '%s'", tarGzURL, rs.Primary.Attributes["url"])
					}

					urlRegex := regexp.MustCompile(`^https?://.+\.tar.gz$`)
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

					resource "utilities_extract_tar_gz" "example2" {
						destination = "%s"
						source      = "%s"
					}
					`, backupDir, tagGzFilePath),
				Check: resource.TestCheckFunc(func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["utilities_extract_tar_gz.example2"]
					if !ok {
						return fmt.Errorf("resource not found: utilities_extract_tar_gz.example2")
					}

					if rs.Primary.Attributes["destination"] != backupDir {
						return fmt.Errorf("expected destination to be '%s', but found '%s'", backupDir, rs.Primary.Attributes["destination"])
					}

					if rs.Primary.Attributes["url"] != "" {
						return fmt.Errorf("expected url to be empty, but found '%s'", rs.Primary.Attributes["url"])
					}

					if rs.Primary.Attributes["source"] != tagGzFilePath {
						return fmt.Errorf("expected source to be '%s', got '%s'", tagGzFilePath, rs.Primary.Attributes["source"])
					}

					return nil
				}),
			},
		},
	})
}
