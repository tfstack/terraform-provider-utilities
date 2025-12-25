// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func Test_http_request_200(t *testing.T) {
	t.Setenv("HTTP_REQ_RETRY_MODE", "false")

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					output "test" {
						value = provider::utilities::http_request("http://httpstat.us", "GET", "", {}).status_code
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "200"),
				),
			},
		},
	})
}

func Test_http_request_404(t *testing.T) {
	t.Setenv("HTTP_REQ_RETRY_MODE", "false")

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
						output "test" {
							value = provider::utilities::http_request("http://httpstat.us/404", "GET", "", {}).status_code
						}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "404"),
				),
			},
		},
	})
}

func Test_http_request_500(t *testing.T) {
	t.Setenv("HTTP_REQ_RETRY_MODE", "false")

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					output "test" {
						value = provider::utilities::http_request("http://httpstat.us/500", "GET", "", {}).status_code
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "500"),
				),
			},
		},
	})
}
