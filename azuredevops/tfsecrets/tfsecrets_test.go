package tfsecrets

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/operations"
	"github.com/stretchr/testify/require"
)

func tetsAccProviders()  {
  
}

func TestAccSvcEp_update(t *testing.T) {
	// var svcEp foo.SvcEp

	resource.Test(t, resource.TestCase{
		// PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
        // use a dynamic configuration with the random name from above
        // setup the svc ednpoint for the first time provide a 
        // PAT. Expect that the PAT is not present, but that a PAT_HASH is.
				Config: testAccResource(rName),
				Check: resource.ComposeTestCheckFunc(
          func (s *terraform.State) error {
            return nil
          },
				),
			},
			{
        // use a dynamic configuration with the random name from above
        // setup the svc endpoint for the 2nd time. Provide the same PAT as before. 
        // expect that the PAT is not present, but that a PAT HASH is.
				Config: testAccResourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
          func (s *terraform.State) error {
            return nil
          },
				),
			},
		},
	})
}

// testAccResource returns an configuration for an  SvcEp with the provided name
func testAccResourceUpdated(name string) string {
	return fmt.Sprintf(`
resource "_SvcEp" "foo" {
  active = false
  name = "%s"
}`, name)
}

func testAccProjectCheckDestroy(s *terraform.State) error {
	clients := testAccProvider.Meta().(*aggregatedClient)

	// verify that every project referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "azuredevops_project" {
			continue
		}

		id := resource.Primary.ID

		// indicates the project still exists - this should fail the test
		if _, err := projectRead(clients, id, ""); err == nil {
			return fmt.Errorf("project with ID %s should not exist", id)
		}
	}

	return nil
}
