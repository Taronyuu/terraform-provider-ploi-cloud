package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_volume.storage", "name", "app-storage"),
					resource.TestCheckResourceAttr("ploicloud_volume.storage", "size", "10"),
					resource.TestCheckResourceAttr("ploicloud_volume.storage", "mount_path", "/var/www/html/storage"),
					resource.TestCheckResourceAttrSet("ploicloud_volume.storage", "id"),
					resource.TestCheckResourceAttrSet("ploicloud_volume.storage", "application_id"),
				),
			},
			{
				ResourceName:      "ploicloud_volume.storage",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1.1",
			},
		},
	})
}

func TestAccVolumeResourceResize(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_volume.storage", "size", "10"),
				),
			},
			{
				Config: testAccVolumeResourceConfigResized(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_volume.storage", "size", "20"),
				),
			},
		},
	})
}

func testAccVolumeResourceConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_volume" "storage" {
  application_id = ploicloud_application.test.id
  name          = "app-storage"
  size          = 10
  mount_path    = "/var/www/html/storage"
}
`
}

func testAccVolumeResourceConfigResized() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_volume" "storage" {
  application_id = ploicloud_application.test.id
  name          = "app-storage"
  size          = 20
  mount_path    = "/var/www/html/storage"
}
`
}