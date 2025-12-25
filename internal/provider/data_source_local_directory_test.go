// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceLocalDirectory(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")

	requireNoError(t, os.Mkdir(subDir, 0755))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfigLocalDirectory(tempDir),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.utilities_local_directory.test", "exists", "true"),
					resource.TestCheckResourceAttr("data.utilities_local_directory.test", "id", tempDir),
					resource.TestCheckResourceAttr("data.utilities_local_directory.test", "path", tempDir),
				),
			},
		},
	})
}

func testConfigLocalDirectory(dir string) string {
	return `
data "utilities_local_directory" "test" {
  path = "` + dir + `"
}`
}

func requireNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
