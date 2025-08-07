package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_service.mysql", "type", "mysql"),
					resource.TestCheckResourceAttr("ploicloud_service.mysql", "version", "8.0"),
					resource.TestCheckResourceAttrSet("ploicloud_service.mysql", "id"),
					resource.TestCheckResourceAttrSet("ploicloud_service.mysql", "application_id"),
				),
			},
			{
				ResourceName:      "ploicloud_service.mysql",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1.1",
			},
		},
	})
}

func TestAccServiceResourceRedis(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceRedisConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_service.redis", "type", "redis"),
					resource.TestCheckResourceAttr("ploicloud_service.redis", "version", "7.0"),
				),
			},
		},
	})
}

func testAccServiceResourceConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_service" "mysql" {
  application_id = ploicloud_application.test.id
  type          = "mysql"
  version       = "8.0"
  
  settings = {
    database = "production"
    size     = "5Gi"
  }
}
`
}

func testAccServiceResourceRedisConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_service" "redis" {
  application_id = ploicloud_application.test.id
  type          = "redis"
  version       = "7.0"
}
`
}