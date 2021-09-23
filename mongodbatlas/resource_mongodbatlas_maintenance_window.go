package mongodbatlas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mwielbut/pointy"
	"github.com/spf13/cast"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

const (
	errorMaintenanceCreate    = "error creating the MongoDB Atlas Maintenance Window (%s): %s"
	errorMaintenanceUpdate    = "error updating the MongoDB Atlas Maintenance Window (%s): %s"
	errorMaintenanceRead      = "error reading the MongoDB Atlas Maintenance Window (%s): %s"
	errorMaintenanceDelete    = "error deleting the MongoDB Atlas Maintenance Window (%s): %s"
	errorMaintenanceDefer     = "error deferring the MongoDB Atlas Maintenance Window (%s): %s"
	errorMaintenanceAutoDefer = "error auto deferring the MongoDB Atlas Maintenance Window (%s): %s"
)

func resourceMongoDBAtlasMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMongoDBAtlasMaintenanceWindowCreate,
		ReadContext:   resourceMongoDBAtlasMaintenanceWindowRead,
		UpdateContext: resourceMongoDBAtlasMaintenanceWindowUpdate,
		DeleteContext: resourceMongoDBAtlasMaintenanceWindowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"day_of_week": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 1 || v > 7 {
						errs = append(errs, fmt.Errorf("%q value should be between 1 and 7, got: %d", key, v))
					}
					return
				},
			},
			"hour_of_day": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"start_asap"},
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 23 {
						errs = append(errs, fmt.Errorf("%q value should be between 0 and 23, got: %d", key, v))
					}
					return
				},
			},
			"start_asap": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"number_of_deferrals": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"defer": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"auto_defer": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"auto_defer_once_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceMongoDBAtlasMaintenanceWindowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client connection.
	conn := meta.(*MongoDBClient).Atlas

	projectID := d.Get("project_id").(string)

	if deferValue := d.Get("defer").(bool); deferValue {
		_, err := conn.MaintenanceWindows.Defer(ctx, projectID)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorMaintenanceDefer, projectID, err))
		}
	}

	if autoDeferValue := d.Get("auto_defer").(bool); autoDeferValue {
		_, err := conn.MaintenanceWindows.AutoDefer(ctx, projectID)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorMaintenanceAutoDefer, projectID, err))
		}
	}

	maintenanceWindowReq := &matlas.MaintenanceWindow{}

	if dayOfWeek, ok := d.GetOk("day_of_week"); ok {
		maintenanceWindowReq.DayOfWeek = cast.ToInt(dayOfWeek)
	}

	if hourOfDay, ok := d.GetOk("hour_of_day"); ok {
		maintenanceWindowReq.HourOfDay = pointy.Int(cast.ToInt(hourOfDay))
	}

	if numberOfDeferrals, ok := d.GetOk("number_of_deferrals"); ok {
		maintenanceWindowReq.NumberOfDeferrals = cast.ToInt(numberOfDeferrals)
	}

	if autoDeferOnceEnabled, ok := d.GetOk("auto_defer_once_enabled"); ok {
		maintenanceWindowReq.AutoDeferOnceEnabled = pointy.Bool(autoDeferOnceEnabled.(bool))
	}

	_, err := conn.MaintenanceWindows.Update(ctx, projectID, maintenanceWindowReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceCreate, projectID, err))
	}

	d.SetId(projectID)

	return resourceMongoDBAtlasMaintenanceWindowRead(ctx, d, meta)
}

func resourceMongoDBAtlasMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client connection.
	conn := meta.(*MongoDBClient).Atlas

	maintenanceWindow, resp, err := conn.MaintenanceWindows.Get(context.Background(), d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf(errorMaintenanceRead, d.Id(), err))
	}

	if err := d.Set("day_of_week", maintenanceWindow.DayOfWeek); err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceRead, d.Id(), err))
	}

	if err := d.Set("hour_of_day", maintenanceWindow.HourOfDay); err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceRead, d.Id(), err))
	}

	if err := d.Set("number_of_deferrals", maintenanceWindow.NumberOfDeferrals); err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceRead, d.Id(), err))
	}
	// start_asap is just display the state of the maintenance,
	// and it doesn't able to set it because breaks the Terraform flow
	// it can be used via API
	if err := d.Set("start_asap", maintenanceWindow.StartASAP); err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceRead, d.Id(), err))
	}

	if maintenanceWindow.AutoDeferOnceEnabled != nil {
		if err := d.Set("auto_defer_once_enabled", *maintenanceWindow.AutoDeferOnceEnabled); err != nil {
			return diag.Errorf(errorMaintenanceRead, d.Id(), err)
		}
	}

	if err := d.Set("project_id", d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceRead, d.Id(), err))
	}

	return nil
}

func resourceMongoDBAtlasMaintenanceWindowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client connection.
	conn := meta.(*MongoDBClient).Atlas

	maintenanceWindowReq := &matlas.MaintenanceWindow{}

	if d.HasChange("defer") {
		_, err := conn.MaintenanceWindows.Defer(ctx, d.Id())
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorMaintenanceDefer, d.Id(), err))
		}
	}

	if d.HasChange("auto_defer") {
		_, err := conn.MaintenanceWindows.AutoDefer(ctx, d.Id())
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorMaintenanceAutoDefer, d.Id(), err))
		}
	}

	if d.HasChange("day_of_week") {
		maintenanceWindowReq.DayOfWeek = cast.ToInt(d.Get("day_of_week"))
	}

	if d.HasChange("hour_of_day") {
		maintenanceWindowReq.HourOfDay = pointy.Int(cast.ToInt(d.Get("hour_of_day")))
	}

	if d.HasChange("number_of_deferrals") {
		maintenanceWindowReq.NumberOfDeferrals = cast.ToInt(d.Get("number_of_deferrals"))
	}

	if d.HasChange("auto_defer_once_enabled") {
		maintenanceWindowReq.AutoDeferOnceEnabled = pointy.Bool(d.Get("number_of_deferrals").(bool))
	}

	_, err := conn.MaintenanceWindows.Update(ctx, d.Id(), maintenanceWindowReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceUpdate, d.Id(), err))
	}

	return nil
}

func resourceMongoDBAtlasMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client connection.
	conn := meta.(*MongoDBClient).Atlas

	_, err := conn.MaintenanceWindows.Reset(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMaintenanceDelete, d.Id(), err))
	}

	return nil
}
