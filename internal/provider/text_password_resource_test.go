// Copyright (c) Plex, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTextPasswordResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTextPasswordResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pwpusher_text.test", "password", "one"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTextPasswordResourceConfig(password string) string {
	return fmt.Sprintf(`
resource "pwpusher_text" "test" {
  password = %[1]q
}
`, password)
}
