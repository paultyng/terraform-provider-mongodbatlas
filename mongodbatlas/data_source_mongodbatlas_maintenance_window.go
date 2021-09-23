package mongodbatlas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spf13/cast"
)

func dataSourceMongoDBAtlasMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasMaintenanceWindowRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"day_of_week": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hour_of_day": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"start_asap": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"number_of_deferrals": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"auto_defer_once_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceMongoDBAtlasMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	projectID := d.Get("project_id").(string)

	maintenance, _, err := conn.MaintenanceWindows.Get(ctx, projectID)
	if err != nil {
		return diag.Errorf(errorMaintenanceRead, projectID, err)
	}

	if err := d.Set("day_of_week", maintenance.DayOfWeek); err != nil {
		return diag.Errorf(errorMaintenanceRead, projectID, err)
	}

	if err := d.Set("hour_of_day", maintenance.HourOfDay); err != nil {
		return diag.Errorf(errorMaintenanceRead, projectID, err)
	}

	if err := d.Set("number_of_deferrals", maintenance.NumberOfDeferrals); err != nil {
		return diag.Errorf(errorMaintenanceRead, projectID, err)
	}

	if err := d.Set("start_asap", cast.ToBool(maintenance.StartASAP)); err != nil {
		return diag.Errorf(errorMaintenanceRead, projectID, err)
	}

	if err := d.Set("auto_defer_once_enabled", cast.ToBool(maintenance.AutoDeferOnceEnabled)); err != nil {
		return diag.Errorf(errorMaintenanceRead, projectID, err)
	}

	d.SetId(projectID)

	return nil
}
