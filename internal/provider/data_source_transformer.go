package provider

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-scaffolding/utils"
)

func dataSourceTransformer() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceTransformerRead,
		Description: "The `file_transformer` data source provides an interface between terraform " +
			"and the file manager of the machine that is running terraform, allowing to overwrite, delete/edit file contents. " +
			"The `file_transformer` data source can be used with existing or non-existing files, " +
			"currently supported file extensions are json, .env and yaml (or yml)" +
			"\n" +
			"**Warning** It is necessary that you grant sufficient permissions so that the provider can read " +
			"and make changes to the contents of the specified file (_chmod +rw_). If the file does not exist, the `file` provider " +
			"will try to create a new file or folder (in this case the file must be placed in the subfolder that does not exist), " +
			"so the permissions must also cover these situations.",
		Schema: map[string]*schema.Schema{
			"file": &schema.Schema{
				Description:  "",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validateFileExt([]string{".json", ".yaml", ".yml", ".env"}),
			},
			"output": &schema.Schema{
				Description:  "c",
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validateFileExt([]string{".json", ".yaml", ".yml"}),
			},
			"override_array_items": &schema.Schema{
				Description: "b",
				Optional:    true,
				Default:     true,
				Type:        schema.TypeBool,
			},
			"items": &schema.Schema{
				Description: "a",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceTransformerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	m := meta.(*utils.Client)

	filePath := d.Get("file").(string)
	items := d.Get("items").(string)
	fileOutputPath := d.Get("output").(string)
	overrideArrayItems := d.Get("override_array_items").(bool)
	//If the outputPath value is not provided, the input filePath value is assigned to the outputPath value
	if fileOutputPath == "" {
		fileOutputPath = filePath
		d.Set("output", filePath)
	}
	err := m.FileTransform(filePath, items, fileOutputPath, utils.WithOverrideArrayItems(overrideArrayItems))
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Query Handling Failed",
				Detail:   err.Error(),
			},
		}
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func validateFileExt(validExt []string) func(v interface{}, s string) ([]string, []error) {
	return func(v interface{}, s string) ([]string, []error) {
		var validExtStr string
		value := v.(string)
		fileExtension := filepath.Ext(value)
		// Verify if the value provided by the user is included in the list of expected values
		for _, v := range validExt {
			if v == fileExtension {
				return nil, nil
			}
		}
		validExtStr = strings.Join(validExt, " ")
		return nil, []error{errors.New(fmt.Sprintf("The file extension is not supported. The supported extensions are the following: %v", validExtStr))}
	}
}
