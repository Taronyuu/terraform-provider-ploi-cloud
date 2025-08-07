package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDomainResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_domain.example", "domain", "example.com"),
					resource.TestCheckResourceAttrSet("ploicloud_domain.example", "id"),
					resource.TestCheckResourceAttrSet("ploicloud_domain.example", "application_id"),
				),
			},
			{
				ResourceName:      "ploicloud_domain.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1.1",
			},
		},
	})
}

func TestAccDomainResourceMultiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceMultipleConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_domain.primary", "domain", "example.com"),
					resource.TestCheckResourceAttr("ploicloud_domain.www", "domain", "www.example.com"),
				),
			},
		},
	})
}

func testAccDomainResourceConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_domain" "example" {
  application_id = ploicloud_application.test.id
  domain        = "example.com"
}
`
}

func testAccDomainResourceMultipleConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_domain" "primary" {
  application_id = ploicloud_application.test.id
  domain        = "example.com"
}

resource "ploicloud_domain" "www" {
  application_id = ploicloud_application.test.id
  domain        = "www.example.com"
}
`
}