package azuredevops

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/stretchr/testify/require"
	"testing"
)

//var testProjectID2 = uuid.New().String()

var testVariableGroup = taskagent.VariableGroup{
	Description: converter.String("Description"),
	Id:          converter.Int(100),
	Name:        converter.String("Name"),
	Type:        converter.String("Vsts"),
	Variables: &map[string]taskagent.VariableValue{
		"var1": {
			Value: converter.String("value1"),
		},
	},
}

// verifies that the flatten/expand round trip yields the same variable group
func TestAzureDevOpsVariableGroup_ExpandFlatten_Roundtrip(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceVariableGroup().Schema, nil)
	flattenVariableGroup(resourceData, &testVariableGroup, testProjectID)

	variableGroupAfterRoundTrip, projectID, err := expandVariableGroup(resourceData, testVariableGroup.Id)

	require.Nil(t, err)
	require.Equal(t, testVariableGroup, *variableGroupAfterRoundTrip)
	require.Equal(t, testProjectID, projectID)
}

// verifies that an expand will fail if there is insufficient configuration data found in the resource
func TestAzureDevOpsVariableGroup_Expand_FailsIfNotEnoughData(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceVariableGroup().Schema, nil)
	_, _, err := expandVariableGroup(resourceData, testVariableGroup.Id)
	require.NotNil(t, err)
}

// TODO : MOCK METHOD

// validates that an apply followed by another apply (i.e., resource update) will be reflected in AzDO and the
// underlying terraform state.
func TestAccAzureDevOpsVariableGroup_CreateAndUpdate(t *testing.T) {
	projectName := testAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	variableGroupNameFirst := testAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	variableGroupNameSecond := testAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	tfVariableGroupNode := "azuredevops_variable_group.vg"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccVariableGroupCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVariableGroupResource(projectName, variableGroupNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfVariableGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVariableGroupNode, "name", variableGroupNameFirst),
					testAccCheckVariableGroupResourceExists(variableGroupNameFirst),
				),
			}, {
				Config: testAccVariableGroupResource(projectName, variableGroupNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfVariableGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVariableGroupNode, "name", variableGroupNameSecond),
					testAccCheckVariableGroupResourceExists(variableGroupNameSecond),
				),
			},
		},
	})
}

// HCL describing an AzDO variable group
func testAccVariableGroupResource(projectName string, variableGroupName string) string {
	variableGroupResource := fmt.Sprintf(`
resource "azuredevops_variable_group" "vg" {
	project_id      = azuredevops_project.project.id
	name            = "%s"
	
  	description = "Description of Sample VG1"

  variables {
    name = "key1"
    value = "value1"
  }
  variables {
    name = "key2"
    value = "value2"
  }

}`, variableGroupName)

	projectResource := testAccProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, variableGroupResource)
}

// Given the name of an AzDO variable group, this will return a function that will check whether
// or not the definition (1) exists in the state and (2) exist in AzDO and (3) has the correct name
func testAccCheckVariableGroupResourceExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		varGroup, ok := s.RootModule().Resources["azuredevops_variable_group.vg"]
		if !ok {
			return fmt.Errorf("Did not find a variable group in the TF state")
		}

		variableGroup, err := getVariableGroupFromResource(varGroup)
		if err != nil {
			return err
		}

		if *variableGroup.Name != expectedName {
			return fmt.Errorf("Variable Group has Name=%s, but expected Name=%s", *variableGroup.Name, expectedName)
		}

		return nil
	}
}

// verifies that all variable groups referenced in the state are destroyed. This will be invoked
// *after* terrafform destroys the resource but *before* the state is wiped clean.
func testAccVariableGroupCheckDestroy(s *terraform.State) error {
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "azuredevops_variable_group" {
			continue
		}

		// indicates the variable group still exists - this should fail the test
		if _, err := getVariableGroupFromResource(resource); err == nil {
			return fmt.Errorf("Unexpectedly found a variable group that should be deleted")
		}
	}

	return nil
}

// given a resource from the state, return a variable group (and error)
func getVariableGroupFromResource(resource *terraform.ResourceState) (*taskagent.VariableGroup, error) {
	//variableGroupID, err := strconv.Atoi(resource.Primary.ID)

	projectID := resource.Primary.Attributes["project_id"]
	clients := testAccProvider.Meta().(*aggregatedClient)

	_, variableGroupID, _ := GetComputedId(clients, resource.Primary.ID)

	return clients.TaskAgentClient.GetVariableGroup(clients.ctx, taskagent.GetVariableGroupArgs{
		Project: &projectID,
		GroupId: &variableGroupID,
	})
}
