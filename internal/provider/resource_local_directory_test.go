package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestResourceUtilitiesLocalDirectory(t *testing.T) {
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
						path        = "%s"
						permissions = "0755"  # Explicitly set permissions to "0755"
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

					if !info.IsDir() {
						return fmt.Errorf("expected a directory but found a file: %s", dirPath1)
					}

					// currentUser, err := user.Current()
					// if err != nil {
					// 	return fmt.Errorf("failed to retrieve current user: %v", err)
					// }
					// userName := currentUser.Username

					// groupObj, groupErr := user.LookupGroupId(currentUser.Gid)
					// groupName := currentUser.Gid // Fallback to GID as a string
					// if groupErr == nil {
					// 	groupName = groupObj.Name
					// }

					// if currentUser.Username != userName || groupObj.Name != groupName {
					// 	return fmt.Errorf("expected user '%s' and group '%s', got user: '%s', group: '%s'", userName, groupName, currentUser.Username, groupObj.Name)
					// }

					permissions := info.Mode().Perm()
					if fmt.Sprintf("%04o", permissions) != "0755" {
						return fmt.Errorf("expected permissions '0755', got: %04o", permissions)
					}

					return nil
				}),
			},
			// {
			// 	Config: fmt.Sprintf(`
			// 		terraform {
			// 			required_providers {
			// 				utilities = {
			// 					source = "hashicorp.com/tfstack/utilities"
			// 				}
			// 			}
			// 		}

			// 		provider "utilities" {}

			// 		resource "utilities_local_directory" "example" {
			// 			force       = true
			// 			group       = "root"
			// 			path        = "%s"
			// 			permissions = "0755"
			// 			user        = "root"
			// 			depends_on = [utilities_local_directory.example]
			// 		}
			// 	`, dirPath2),
			// 	Destroy: true,
			// 	Check: resource.TestCheckFunc(func(s *terraform.State) error {
			// 		if _, err := os.Stat(dirPath2); !os.IsNotExist(err) {
			// 			return fmt.Errorf("expected directory %s to be removed after destroy, but it still exists", dirPath2)
			// 		}
			// 		return nil
			// 	}),
			// },
		},
	})
}
