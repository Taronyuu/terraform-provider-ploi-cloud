package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecretResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_secret.app_key", "key", "APP_KEY"),
					resource.TestCheckResourceAttr("ploicloud_secret.app_key", "value", "base64:test-key"),
					resource.TestCheckResourceAttrSet("ploicloud_secret.app_key", "application_id"),
				),
			},
			{
				ResourceName:            "ploicloud_secret.app_key",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "1.APP_KEY",
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func TestAccSecretResourceUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_secret.app_key", "value", "base64:test-key"),
				),
			},
			{
				Config: testAccSecretResourceConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_secret.app_key", "value", "base64:updated-key"),
				),
			},
		},
	})
}

func testAccSecretResourceConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_secret" "app_key" {
  application_id = ploicloud_application.test.id
  key           = "APP_KEY"
  value         = "base64:test-key"
}
`
}

func testAccSecretResourceConfigUpdated() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_secret" "app_key" {
  application_id = ploicloud_application.test.id
  key           = "APP_KEY"
  value         = "base64:updated-key"
}
`
}