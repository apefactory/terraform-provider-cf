package cloudfoundry

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const appResourceSpringMusic = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}
data "cloudfoundry_service" "mysql" {
    name = "p-mysql"
}
data "cloudfoundry_service" "rmq" {
    name = "p-rabbitmq"
}

resource "cloudfoundry_route" "spring-music" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "spring-music"
}
resource "cloudfoundry_service_instance" "db" {
	name = "db"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.mysql.service_plans.512mb}"
}
resource "cloudfoundry_service_instance" "fs1" {
	name = "fs1"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.rmq.service_plans.standard}"
}
resource "cloudfoundry_app" "spring-music" {
	name = "spring-music"
	space = "${data.cloudfoundry_space.space.id}"
	memory = "768"
	disk_quota = "512"
	timeout = 1800

	url = "https://github.com/mevansam/spring-music/releases/download/v1.0/spring-music.war"

	service_binding {
		service_instance = "${cloudfoundry_service_instance.db.id}"
	}
	service_binding {
		service_instance = "${cloudfoundry_service_instance.fs1.id}"
	}

	route {
		default_route = "${cloudfoundry_route.spring-music.id}"
	}

	environment {
		TEST_VAR_1 = "testval1"
		TEST_VAR_2 = "testval2"
	}
}
`

const appResourceSpringMusicUpdate = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}
data "cloudfoundry_service" "mysql" {
    name = "p-mysql"
}
data "cloudfoundry_service" "rmq" {
    name = "p-rabbitmq"
}

resource "cloudfoundry_route" "spring-music" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "spring-music"
}
resource "cloudfoundry_service_instance" "db" {
	name = "db"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.mysql.service_plans.512mb}"
}
resource "cloudfoundry_service_instance" "fs1" {
	name = "fs1"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.rmq.service_plans.standard}"
}
resource "cloudfoundry_service_instance" "fs2" {
	name = "fs2"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.rmq.service_plans.standard}"
}
resource "cloudfoundry_app" "spring-music" {
	name = "spring-music-updated"
	space = "${data.cloudfoundry_space.space.id}"
	instances ="2"
	memory = "1024"
	disk_quota = "1024"
	timeout = 1800

	url = "https://github.com/mevansam/spring-music/releases/download/v1.0/spring-music.war"

	service_binding {
		service_instance = "${cloudfoundry_service_instance.db.id}"
	}
	service_binding {
		service_instance = "${cloudfoundry_service_instance.fs2.id}"
	}
	service_binding {
		service_instance = "${cloudfoundry_service_instance.fs1.id}"
	}

	route {
		default_route = "${cloudfoundry_route.spring-music.id}"
	}

	environment {
		TEST_VAR_1 = "testval1"
		TEST_VAR_2 = "testval2"
	}
}
`

const appResourceWithMultiplePorts = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_app" "test-app" {
	name = "test-app"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 1800
	ports = [ 8888, 9999 ]
	buildpack = "binary_buildpack"
	command = "chmod 0755 test-app && ./test-app --ports=8888,9999"
	health_check_type = "process"

	github_release {
		owner = "mevansam"
		repo = "test-app"
		filename = "test-app"
		version = "v0.0.1"
		user = "%s"
		password = "%s"
	}
}
resource "cloudfoundry_route" "test-app-8888" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app-8888"

	target {
		app = "${cloudfoundry_app.test-app.id}"
		port = 8888
	}
}
resource "cloudfoundry_route" "test-app-9999" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app-9999"

	target {
		app = "${cloudfoundry_app.test-app.id}"
		port = 9999
	}
}
`

const appResourceDocker = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_route" "test-docker-app" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "test-docker-app"
	target {
		app = "${cloudfoundry_app.test-docker-app.id}"
		port = 8080
	}
}
resource "cloudfoundry_app" "test-docker-app" {
	name = "test-docker-app"
	space = "${data.cloudfoundry_space.space.id}"
	docker_image = "cloudfoundry/diego-docker-app:latest"
	timeout = 900
}

`

const multipleVersion = `
data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}
resource "cloudfoundry_route" "test-app" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app" 
    target = {app = "${cloudfoundry_app.test-app.id}"}
}
resource "cloudfoundry_app" "test-app" {
	name = "test-app"
	space = "${data.cloudfoundry_space.space.id}"
	command = "test-app --ports=8080"
	timeout = 1800
    memory = "512"
	git {
		url = "https://github.com/mevansam/test-app.git"
	}
}
`

const multipleVersionUpdate = `
data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_route" "test-app" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app"
    target = {app = "${cloudfoundry_app.test-app.id}"}
}
resource "cloudfoundry_app" "test-app" {
	name = "test-app"
	space = "${data.cloudfoundry_space.space.id}"
	command = "test-app --ports=8080"
	timeout = 1800
    memory = "1024"
	git {
		url = "https://github.com/janosbinder/test-app.git"
	}
}
`

const createManyJavaSpringApps = `

data "cloudfoundry_domain" "java-spring-domain" {
    name = "%s"
}

data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_route" "java-spring-route-1" {
	domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
    space = "${data.cloudfoundry_space.space.id}"
	hostname = "java-spring-1"
	depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-1" {
    name = "java-spring-app-1"
	url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 700
    memory = 512
    buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

	route {
		default_route = "${cloudfoundry_route.java-spring-route-1.id}"
	}

	depends_on = ["cloudfoundry_route.java-spring-route-1"]
}

resource "cloudfoundry_route" "java-spring-route-2" {
	domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
    space = "${data.cloudfoundry_space.space.id}"
	hostname = "java-spring-2"
	depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-2" {
    name = "java-spring-app-2"
	url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 700
    memory = 512
    buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

	route {
		default_route = "${cloudfoundry_route.java-spring-route-2.id}"
	}

	depends_on = ["cloudfoundry_route.java-spring-route-2"]
}

resource "cloudfoundry_route" "java-spring-route-3" {
	domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
    space = "${data.cloudfoundry_space.space.id}"
	hostname = "java-spring-3"
	depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-3" {
    name = "java-spring-app-3"
	url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 700
    memory = 512
    buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

	route {
		default_route = "${cloudfoundry_route.java-spring-route-3.id}"
	}

	depends_on = ["cloudfoundry_route.java-spring-route-3"]
}

resource "cloudfoundry_route" "java-spring-route-4" {
	domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
    space = "${data.cloudfoundry_space.space.id}"
	hostname = "java-spring-4"
	depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-4" {
    name = "java-spring-app-4"
	url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 700
    memory = 512
    buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

	route {
		default_route = "${cloudfoundry_route.java-spring-route-4.id}"
	}

	depends_on = ["cloudfoundry_route.java-spring-route-4"]
}

resource "cloudfoundry_route" "java-spring-route-5" {
	domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
    space = "${data.cloudfoundry_space.space.id}"
	hostname = "java-spring-5"
	depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-5" {
    name = "java-spring-app-5"
	url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 700
    memory = 512
    buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

	route {
		default_route = "${cloudfoundry_route.java-spring-route-5.id}"
	}

	depends_on = ["cloudfoundry_route.java-spring-route-5"]
}
`

// If the PR is not applied, after running this test many times, it should crash with this error
// === RUN   TestAccApp_reproduceIssue88
// Application downloaded to: ../tests/cf-acceptance-tests/assets/java-spring/
// Application downloaded to: ../tests/cf-acceptance-tests/assets/java-spring/
// fatal error: concurrent map read and map write
//
// goroutine 1542 [running]:
// ...
// created by github.com/terraform-providers/terraform-provider-cf/cloudfoundry.resourceAppCreate
// .../golang/src/github.com/terraform-providers/terraform-provider-cf/cloudfoundry/resource_cf_app.go:421 +0x1ac4

func TestAccApp_reproduceIssue88(t *testing.T) {
	refApp1 := "cloudfoundry_app.java-spring-app-1"
	refApp2 := "cloudfoundry_app.java-spring-app-2"
	refApp3 := "cloudfoundry_app.java-spring-app-3"
	refApp4 := "cloudfoundry_app.java-spring-app-4"
	refApp5 := "cloudfoundry_app.java-spring-app-5"

	failRegExp, _ := regexp.Compile("app java-spring-app-[0-9] failed to start")

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring-app-`", "java-spring-app-2", "java-spring-app-3", "java-spring-app-4", "java-spring-app-5"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(createManyJavaSpringApps, defaultAppDomain()),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckAppExists(refApp1, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-1."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp2, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp3, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-3."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp4, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-4."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp5, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-5."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
					// the jar in the test is enough big, and allows us to test for the failure
					ExpectError: failRegExp,
				},
			},
		})
}

func TestAccAppVersions_app1(t *testing.T) {

	refRoute := "cloudfoundry_route.test-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"test-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(multipleVersion, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {

							if err = assertHTTPResponse("https://test-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(multipleVersionUpdate, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {

							if err = assertHTTPResponse("https://test-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
				},
			},
		})
}

func TestAccApp_app1(t *testing.T) {

	refApp := "cloudfoundry_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceSpringMusic, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(
							refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(
							refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(
							refApp, "instances", "1"),
						resource.TestCheckResourceAttr(
							refApp, "memory", "768"),
						resource.TestCheckResourceAttr(
							refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(
							refApp, "stack"),
						resource.TestCheckResourceAttr(
							refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(
							refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(
							refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(
							refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(
							refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(
							refApp, "service_binding.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(appResourceSpringMusicUpdate, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refApp, "name", "spring-music-updated"),
						resource.TestCheckResourceAttr(
							refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(
							refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(
							refApp, "instances", "2"),
						resource.TestCheckResourceAttr(
							refApp, "memory", "1024"),
						resource.TestCheckResourceAttr(
							refApp, "disk_quota", "1024"),
						resource.TestCheckResourceAttrSet(
							refApp, "stack"),
						resource.TestCheckResourceAttr(
							refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(
							refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(
							refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(
							refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(
							refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(
							refApp, "service_binding.#", "3"),
					),
				},
			},
		})
}
func TestAccApp_app2(t *testing.T) {

	refApp := "cloudfoundry_app.test-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"test-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceWithMultiplePorts, defaultAppDomain(), os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_TOKEN")),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {
							responses := []string{"8888"}
							if err = assertHTTPResponse("https://test-app-8888."+defaultAppDomain()+"/port", 200, &responses); err != nil {
								return err
							}
							responses = []string{"9999"}
							if err = assertHTTPResponse("https://test-app-9999."+defaultAppDomain()+"/port", 200, &responses); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refApp, "name", "test-app"),
						resource.TestCheckResourceAttr(
							refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "2"),
						resource.TestCheckResourceAttr(
							refApp, "ports.8888", "8888"),
						resource.TestCheckResourceAttr(
							refApp, "ports.9999", "9999"),
					),
				},
			},
		})
}

func TestAccApp_dockerApp(t *testing.T) {
	refApp := "cloudfoundry_app.test-docker-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"test-docker-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceDocker, defaultAppDomain()),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://test-docker-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refApp, "name", "test-docker-app"),
						resource.TestCheckResourceAttr(
							refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(
							refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(
							refApp, "instances", "1"),
						resource.TestCheckResourceAttrSet(
							refApp, "stack"),
						resource.TestCheckResourceAttr(
							refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(
							refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(
							refApp, "docker_image", "cloudfoundry/diego-docker-app:latest"),
					),
				},
			},
		})
}

func testAccCheckAppExists(resApp string, validate func() error) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resApp, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var (
			app             cfapi.CCApp
			routeMappings   []map[string]interface{}
			serviceBindings []map[string]interface{}
		)

		am := session.AppManager()
		rm := session.RouteManager()

		if app, err = am.ReadApp(id); err != nil {
			return err
		}
		session.Log.DebugMessage(
			"retrieved app for resource '%s' with id '%s': %# v",
			resApp, id, app)

		if err = assertEquals(attributes, "name", app.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "space", app.SpaceGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "instances", app.Instances); err != nil {
			return err
		}
		if err = assertEquals(attributes, "memory", app.Memory); err != nil {
			return err
		}
		if err = assertEquals(attributes, "disk_quota", app.DiskQuota); err != nil {
			return err
		}
		if err = assertEquals(attributes, "stack", app.StackGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "buildpack", app.Buildpack); err != nil {
			return err
		}
		if err = assertEquals(attributes, "command", app.Command); err != nil {
			return err
		}
		if err = assertEquals(attributes, "enable_ssh", app.EnableSSH); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_http_endpoint", app.HealthCheckHTTPEndpoint); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_type", app.HealthCheckType); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_timeout", app.HealthCheckTimeout); err != nil {
			return err
		}
		if err = assertMapEquals("environment", attributes, *app.Environment); err != nil {
			return err
		}

		if serviceBindings, err = am.ReadServiceBindingsByApp(id); err != nil {
			return err
		}
		session.Log.DebugMessage(
			"retrieved service bindings for app with id '%s': %# v",
			id, serviceBindings)

		if err = assertListEquals(attributes, "service_binding", len(serviceBindings),
			func(values map[string]string, i int) (match bool) {
				var binding map[string]interface{}

				serviceInstanceID := values["service_instance"]
				binding = nil

				for _, b := range serviceBindings {
					if serviceInstanceID == b["service_instance"] {
						binding = b
						break
					}
				}

				if binding != nil && values["binding_id"] == binding["binding_id"] {
					return true
				}
				return false

			}); err != nil {
			return err
		}

		if routeMappings, err = rm.ReadRouteMappingsByApp(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved routes for app with id '%s': %# v",
			id, routeMappings)

		for _, r := range []string{
			"default_route",
			"stage_route",
			"live_route",
		} {
			if err = validateRouteMapping(r, attributes, routeMappings); err != nil {
				return
			}
		}

		err = validate()
		return
	}
}

func validateRouteMapping(routeName string, attributes map[string]string, routeMappings []map[string]interface{}) (err error) {

	var (
		routeID, mappingID string
		mapping            map[string]interface{}

		ok bool
	)

	routeKey := "route.0." + routeName
	routeMappingKey := "route.0." + routeName + "_mapping_id"

	if routeID, ok = attributes[routeKey]; ok && len(routeID) > 0 {
		if mappingID, ok = attributes[routeMappingKey]; !ok || len(mappingID) == 0 {
			return fmt.Errorf("default route '%s' does not have a corresponding mapping id in the state", routeID)
		}

		mapping = nil
		for _, r := range routeMappings {
			if mappingID == r["mapping_id"] {
				mapping = r
				break
			}
		}
		if mapping == nil {
			return fmt.Errorf("unable to find route mapping with id '%s' for route '%s'", mappingID, routeID)
		}
		if routeID != mapping["route"] {
			return fmt.Errorf("route mapping with id '%s' does not map to route '%s'", mappingID, routeID)
		}
	}
	return err
}

func testAccCheckAppDestroyed(apps []string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		for _, a := range apps {
			if _, err := session.AppManager().FindApp(a); err != nil {
				switch err.(type) {
				case *errors.ModelNotFoundError:
					continue
				default:
					return err
				}
			}
			return fmt.Errorf("app with name '%s' still exists in cloud foundry", a)
		}
		return nil
	}
}
