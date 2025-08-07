package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_worker.queue", "name", "queue-worker"),
					resource.TestCheckResourceAttr("ploicloud_worker.queue", "command", "php artisan queue:work --sleep=3"),
					resource.TestCheckResourceAttr("ploicloud_worker.queue", "replicas", "2"),
					resource.TestCheckResourceAttrSet("ploicloud_worker.queue", "id"),
					resource.TestCheckResourceAttrSet("ploicloud_worker.queue", "application_id"),
				),
			},
			{
				ResourceName:      "ploicloud_worker.queue",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1.1",
			},
		},
	})
}

func TestAccWorkerResourceUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_worker.queue", "replicas", "2"),
				),
			},
			{
				Config: testAccWorkerResourceConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_worker.queue", "replicas", "3"),
					resource.TestCheckResourceAttr("ploicloud_worker.queue", "command", "php artisan queue:work --sleep=1 --tries=3"),
				),
			},
		},
	})
}

func testAccWorkerResourceConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_worker" "queue" {
  application_id = ploicloud_application.test.id
  name          = "queue-worker"
  command       = "php artisan queue:work --sleep=3"
  replicas      = 2
}
`
}

func testAccWorkerResourceConfigUpdated() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-app"
  type = "laravel"
}

resource "ploicloud_worker" "queue" {
  application_id = ploicloud_application.test.id
  name          = "queue-worker"
  command       = "php artisan queue:work --sleep=1 --tries=3"
  replicas      = 3
}
`
}