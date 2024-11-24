package provider

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"syscall"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestResourceUtilitiesLocalDirectory(t *testing.T) {
	if isWindows() {
		t.Skip("Windows is not supported, skipping test.")
	}

	dirPath1 := fmt.Sprintf("/tmp/test-dir-%s", uuid.NewString())
	dirPath2 := fmt.Sprintf("/tmp/test-dir-%s", uuid.NewString())

	// Ensure cleanup even on test failure
	defer func() {
		os.RemoveAll(dirPath1)
		os.RemoveAll(dirPath2)
	}()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
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

					resource "utilities_local_directory" "example" {
						force       = true
						group       = "root"
						path        = "%s"
						permissions = "0755"  # Explicitly set permissions to "0755"
						user        = "root"
					}
				`, dirPath1),
				Check: resource.TestCheckFunc(func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["utilities_local_directory.example"]
					if !ok {
						return fmt.Errorf("resource not found: utilities_local_directory.example")
					}

					if rs.Primary.Attributes["path"] != dirPath1 {
						return fmt.Errorf("expected path to be '%s', got '%s'", dirPath1, rs.Primary.Attributes["path"])
					}

					info, err := os.Stat(dirPath1)
					if err != nil {
						return fmt.Errorf("failed to stat directory: %v", err)
					}

					sysInfo, ok := info.Sys().(*syscall.Stat_t)
					if !ok {
						return fmt.Errorf("failed to retrieve system info for directory")
					}

					userObj, err := user.LookupId(fmt.Sprintf("%d", sysInfo.Uid))
					userName := fmt.Sprintf("%d", sysInfo.Uid)
					if err == nil {
						userName = userObj.Username
					}

					groupName, err := getGroupNameByGID(int(sysInfo.Gid))
					if err != nil {
						groupName = fmt.Sprintf("%d", sysInfo.Gid)
					}

					if userName != "root" || groupName != "root" {
						return fmt.Errorf("expected user 'root' and group 'root', got user: '%s', group: '%s'", userName, groupName)
					}

					permissions := info.Mode().Perm()
					if fmt.Sprintf("%04o", permissions) != "0755" {
						return fmt.Errorf("expected permissions '0755', got: %04o", permissions)
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

					resource "utilities_local_directory" "example" {
						force       = true
						group       = "root"
						path        = "%s"
						permissions = "0755"
						user        = "root"
						depends_on = [utilities_local_directory.example]
					}
				`, dirPath2),
				Destroy: true,
				Check: resource.TestCheckFunc(func(s *terraform.State) error {
					if _, err := os.Stat(dirPath2); !os.IsNotExist(err) {
						return fmt.Errorf("expected directory %s to be removed after destroy, but it still exists", dirPath2)
					}
					return nil
				}),
			},
		},
	})
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

func getGroupNameByGID(gid int) (string, error) {
	cmd := exec.Command("getent", "group", fmt.Sprintf("%d", gid))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get group name for GID %d: %v", gid, err)
	}

	group := strings.Split(string(output), ":")[0]
	return group, nil
}
