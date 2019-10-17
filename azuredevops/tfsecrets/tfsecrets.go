package tfsecrets

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/update_check"
)

func DiffSuppresFunc(key, old, new string, d *schema.ResourceData) bool {
	// use `update_check` to calc if there's a diff
	// store `update_checks` memo in `key + "_memo"`
	memoKey := key + "_memo"
	memoIf := d.Get(memoKey)
	memo := fmt.Sprintf("%v", memoIf)
	isUpdating, newMemo, err := update_check.IsUpdating(new, memo)
	return isUpdating
}
