package mongodbatlas

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spf13/cast"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func TestAccResourceMongoDBAtlasMaintenanceWindow_basic(t *testing.T) {
	var (
		maintenance      matlas.MaintenanceWindow
		resourceName     = "mongodbatlas_maintenance_window.test"
		projectID        = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		dayOfWeek        = 7
		hourOfDay        = 3
		dayOfWeekUpdated = 4
		hourOfDayUpdated = 5
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasMaintenanceWindowConfig(projectID, dayOfWeek, hourOfDay),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasMaintenanceWindowExists(resourceName, &maintenance),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "hour_of_day"),

					resource.TestCheckResourceAttr(resourceName, "day_of_week", cast.ToString(dayOfWeek)),
					resource.TestCheckResourceAttr(resourceName, "hour_of_day", cast.ToString(hourOfDay)),
					resource.TestCheckResourceAttr(resourceName, "number_of_deferrals", "0"),

					testAccCheckMongoDBAtlasMaintenanceWindowAttributes("day_of_week", dayOfWeek, &maintenance.DayOfWeek),
				),
			},
			{
				Config: testAccMongoDBAtlasMaintenanceWindowConfig(projectID, dayOfWeekUpdated, hourOfDayUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasMaintenanceWindowExists(resourceName, &maintenance),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "hour_of_day"),

					resource.TestCheckResourceAttr(resourceName, "day_of_week", cast.ToString(dayOfWeekUpdated)),
					resource.TestCheckResourceAttr(resourceName, "hour_of_day", cast.ToString(hourOfDayUpdated)),
					resource.TestCheckResourceAttr(resourceName, "number_of_deferrals", "0"),

					testAccCheckMongoDBAtlasMaintenanceWindowAttributes("day_of_week", dayOfWeekUpdated, &maintenance.DayOfWeek),
				),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasMaintenanceWindow_importBasic(t *testing.T) {
	var (
		maintenance  matlas.MaintenanceWindow
		resourceName = "mongodbatlas_maintenance_window.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		dayOfWeek    = 1
		hourOfDay    = 3
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasMaintenanceWindowConfig(projectID, dayOfWeek, hourOfDay),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasMaintenanceWindowExists(resourceName, &maintenance),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "hour_of_day"),

					resource.TestCheckResourceAttr(resourceName, "day_of_week", cast.ToString(dayOfWeek)),
					resource.TestCheckResourceAttr(resourceName, "hour_of_day", cast.ToString(hourOfDay)),
					resource.TestCheckResourceAttr(resourceName, "number_of_deferrals", "0"),
					resource.TestCheckResourceAttr(resourceName, "start_asap", "false"),

					testAccCheckMongoDBAtlasMaintenanceWindowAttributes("day_of_week", dayOfWeek, &maintenance.DayOfWeek),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckMongoDBAtlasMaintenanceWindowImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceMongoDBAtlasMaintenanceWindow_autoDeferActivated(t *testing.T) {
	var (
		maintenance  matlas.MaintenanceWindow
		resourceName = "mongodbatlas_maintenance_window.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		dayOfWeek    = 7
		hourOfDay    = 3
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasMaintenanceWindowConfigAutoDeferEnabled(projectID, dayOfWeek, hourOfDay),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasMaintenanceWindowExists(resourceName, &maintenance),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "hour_of_day"),

					resource.TestCheckResourceAttr(resourceName, "day_of_week", cast.ToString(dayOfWeek)),
					resource.TestCheckResourceAttr(resourceName, "hour_of_day", cast.ToString(hourOfDay)),
					resource.TestCheckResourceAttr(resourceName, "number_of_deferrals", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_defer_once_enabled", "true"),

					testAccCheckMongoDBAtlasMaintenanceWindowAttributes("day_of_week", dayOfWeek, &maintenance.DayOfWeek),
				),
			},
		},
	})
}

func testAccMongoDBAtlasMaintenanceWindowConfigAutoDeferEnabled(projectID string, dayOfWeek, hourOfDay int) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_maintenance_window" "test" {
			project_id  = "%s"
			day_of_week = %d
			hour_of_day = %d
			auto_defer_once_enabled = true
		}`, projectID, dayOfWeek, hourOfDay)
}

func testAccCheckMongoDBAtlasMaintenanceWindowExists(resourceName string, maintenance *matlas.MaintenanceWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*MongoDBClient).Atlas

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		log.Printf("[DEBUG] projectID: %s", rs.Primary.ID)

		maintenanceWindow, _, err := conn.MaintenanceWindows.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("maintenance Window (%s) does not exist", rs.Primary.ID)
		}

		*maintenance = *maintenanceWindow

		return nil
	}
}

func testAccCheckMongoDBAtlasMaintenanceWindowAttributes(attr string, expected int, got *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if diff := deep.Equal(expected, *got); diff != nil {
			return fmt.Errorf("bad %s \n got = %#v\nwant = %#v \ndiff = %#v", attr, expected, *got, diff)
		}

		return nil
	}
}

func testAccCheckMongoDBAtlasMaintenanceWindowDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*MongoDBClient).Atlas

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mongodbatlas_maintenance_window" {
			continue
		}

		_, _, err := conn.MaintenanceWindows.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("maintenance Window (%s) does not exist", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckMongoDBAtlasMaintenanceWindowImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}

func testAccMongoDBAtlasMaintenanceWindowConfig(projectID string, dayOfWeek, hourOfDay int) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_maintenance_window" "test" {
			project_id  = "%s"
			day_of_week = %d
			hour_of_day = %d
		}`, projectID, dayOfWeek, hourOfDay)
}
