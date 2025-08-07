package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccApplicationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationResourceConfig("test-laravel-app"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_application.test", "name", "test-laravel-app"),
					resource.TestCheckResourceAttr("ploicloud_application.test", "type", "laravel"),
					resource.TestCheckResourceAttrSet("ploicloud_application.test", "id"),
					resource.TestCheckResourceAttrSet("ploicloud_application.test", "url"),
				),
			},
			{
				ResourceName:      "ploicloud_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationResourceConfig("test-laravel-app-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_application.test", "name", "test-laravel-app-updated"),
				),
			},
		},
	})
}

func TestAccApplicationResourceWithRuntime(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationResourceWithRuntimeConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_application.test", "runtime.0.php_version", "8.4"),
					resource.TestCheckResourceAttr("ploicloud_application.test", "runtime.0.nodejs_version", "22"),
				),
			},
		},
	})
}

func TestAccApplicationResourceWithSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationResourceWithSettingsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ploicloud_application.test", "settings.0.replicas", "3"),
					resource.TestCheckResourceAttr("ploicloud_application.test", "settings.0.scheduler_enabled", "true"),
					resource.TestCheckResourceAttr("ploicloud_application.test", "settings.0.health_check_path", "/health"),
				),
			},
		},
	})
}

func testAccApplicationResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "%s"
  type = "laravel"
  
  build_commands = [
    "composer install",
    "npm run build"
  ]
  
  init_commands = [
    "php artisan migrate --force"
  ]
}
`, name)
}

func testAccApplicationResourceWithRuntimeConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-runtime-app"
  type = "laravel"
  
  runtime {
    php_version    = "8.4"
    nodejs_version = "22"
  }
}
`
}

func testAccApplicationResourceWithSettingsConfig() string {
	return `
provider "ploicloud" {
  api_token = "test-token"
}

resource "ploicloud_application" "test" {
  name = "test-settings-app"
  type = "laravel"
  
  settings {
    replicas           = 3
    scheduler_enabled  = true
    health_check_path  = "/health"
    cpu_request       = "1"
    memory_request    = "2Gi"
  }
}
`
}

func testAccCheckApplicationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Application ID is set")
		}

		return nil
	}
}