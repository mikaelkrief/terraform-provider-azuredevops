package azuredevops

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/terraform-provider-azuredevops/azdosdkmocks"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/microsoft/azure-devops-go-api/azuredevops/serviceendpoint"
)

var testServiceEndpointID = uuid.New()
var randomServiceEndpointProjectID = uuid.New().String()
var testServiceEndpointProjectID = &randomServiceEndpointProjectID

var testServiceEndpoint = serviceendpoint.ServiceEndpoint{
	Authorization: &serviceendpoint.EndpointAuthorization{
		Parameters: &map[string]string{
			"accessToken": "UNIT_TEST_ACCESS_TOKEN",
		},
		Scheme: converter.String("PersonalAccessToken"),
	},
	// Description: converter.String("UNIT_TEST_DESCRIPTION"),
	Id: &testServiceEndpointID,
	// IsShared:    converter.Bool(false),
	Name:  converter.String("UNIT_TEST_NAME"),
	Owner: converter.String("library"), // Supported values are "library", "agentcloud"
	Type:  converter.String("UNIT_TEST_TYPE"),
	Url:   converter.String("UNIT_TEST_URL"),
}

/**
 * Begin unit tests
 */

// verifies that the flatten/expand round trip yields the same build definition
func TestAzureDevOpsServiceEndpoint_ExpandFlatten_Roundtrip(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceServiceEndpoint().Schema, nil)
	flattenServiceEndpoint(resourceData, &testServiceEndpoint, testServiceEndpointProjectID)

	serviceEndpointAfterRoundTrip, projectID := expandServiceEndpoint(resourceData)

	require.Equal(t, testServiceEndpoint, *serviceEndpointAfterRoundTrip)
	require.Equal(t, testServiceEndpointProjectID, projectID)
}

// verifies that if an error is produced on create, the error is not swallowed
func TestAzureDevOpsServiceEndpoint_Create_DoesNotSwallowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceData := schema.TestResourceDataRaw(t, resourceServiceEndpoint().Schema, nil)
	flattenServiceEndpoint(resourceData, &testServiceEndpoint, testServiceEndpointProjectID)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &aggregatedClient{ServiceEndpointClient: buildClient, ctx: context.Background()}

	expectedArgs := serviceendpoint.CreateServiceEndpointArgs{Endpoint: &testServiceEndpoint, Project: testServiceEndpointProjectID}
	buildClient.
		EXPECT().
		CreateServiceEndpoint(clients.ctx, expectedArgs).
		Return(nil, errors.New("CreateServiceEndpoint() Failed")).
		Times(1)

	err := resourceServiceEndpointCreate(resourceData, clients)
	require.Equal(t, "Error creating service endpoint in Azure DevOps: CreateServiceEndpoint() Failed", err.Error())
}

// verifies that if an error is produced on a read, it is not swallowed
func TestAzureDevOpsServiceEndpoint_Read_DoesNotSwallowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceData := schema.TestResourceDataRaw(t, resourceServiceEndpoint().Schema, nil)
	flattenServiceEndpoint(resourceData, &testServiceEndpoint, testServiceEndpointProjectID)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &aggregatedClient{ServiceEndpointClient: buildClient, ctx: context.Background()}

	expectedArgs := serviceendpoint.GetServiceEndpointDetailsArgs{EndpointId: testServiceEndpoint.Id, Project: testServiceEndpointProjectID}
	buildClient.
		EXPECT().
		GetServiceEndpointDetails(clients.ctx, expectedArgs).
		Return(nil, errors.New("GetServiceEndpoint() Failed")).
		Times(1)

	err := resourceServiceEndpointRead(resourceData, clients)
	require.Equal(t, "GetServiceEndpoint() Failed", err.Error())
}

// verifies that if an error is produced on a delete, it is not swallowed
func TestAzureDevOpsServiceEndpoint_Delete_DoesNotSwallowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceData := schema.TestResourceDataRaw(t, resourceServiceEndpoint().Schema, nil)
	flattenServiceEndpoint(resourceData, &testServiceEndpoint, testServiceEndpointProjectID)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &aggregatedClient{ServiceEndpointClient: buildClient, ctx: context.Background()}

	expectedArgs := serviceendpoint.DeleteServiceEndpointArgs{EndpointId: testServiceEndpoint.Id, Project: testServiceEndpointProjectID}
	buildClient.
		EXPECT().
		DeleteServiceEndpoint(clients.ctx, expectedArgs).
		Return(errors.New("DeleteServiceEndpoint() Failed")).
		Times(1)

	err := resourceServiceEndpointDelete(resourceData, clients)
	require.Equal(t, "DeleteServiceEndpoint() Failed", err.Error())
}

// verifies that if an error is produced on an update, it is not swallowed
func TestAzureDevOpsServiceEndpoint_Update_DoesNotSwallowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceData := schema.TestResourceDataRaw(t, resourceServiceEndpoint().Schema, nil)
	flattenServiceEndpoint(resourceData, &testServiceEndpoint, testServiceEndpointProjectID)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &aggregatedClient{ServiceEndpointClient: buildClient, ctx: context.Background()}

	expectedArgs := serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   &testServiceEndpoint,
		EndpointId: testServiceEndpoint.Id,
		Project:    testServiceEndpointProjectID,
	}

	buildClient.
		EXPECT().
		UpdateServiceEndpoint(clients.ctx, expectedArgs).
		Return(nil, errors.New("UpdateServiceEndpoint() Failed")).
		Times(1)

	err := resourceServiceEndpointUpdate(resourceData, clients)
	require.Equal(t, "Error updating service endpoint in Azure DevOps: UpdateServiceEndpoint() Failed", err.Error())
}

/**
 * Begin acceptance tests
 */

// // validates that an apply followed by another apply (i.e., resource update) will be reflected in AzDO and the
// // underlying terraform state.
// func TestAccAzureDevOpsBuildDefinition_CreateAndUpdate(t *testing.T) {
// 	projectName := testAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
// 	buildDefinitionNameFirst := testAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
// 	buildDefinitionNameSecond := testAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

// 	tfBuildDefNode := "azuredevops_build_definition.build"
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccBuildDefinitionCheckDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccBuildDefinitionResource(projectName, buildDefinitionNameFirst),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
// 					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
// 					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameFirst),
// 					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameFirst),
// 				),
// 			}, {
// 				Config: testAccBuildDefinitionResource(projectName, buildDefinitionNameSecond),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
// 					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
// 					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameSecond),
// 					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameSecond),
// 				),
// 			},
// 		},
// 	})
// }

// // HCL describing an AzDO build definition
// func testAccBuildDefinitionResource(projectName string, buildDefinitionName string) string {
// 	buildDefinitionResource := fmt.Sprintf(`
// resource "azuredevops_build_definition" "build" {
// 	project_id      = azuredevops_project.project.id
// 	name            = "%s"
// 	agent_pool_name = "Hosted Ubuntu 1604"

// 	repository {
// 	  repo_type             = "GitHub"
// 	  repo_name             = "repoOrg/repoName"
// 	  branch_name           = "branch"
// 	  yml_path              = "path/to/yaml"
// 	}
// }`, buildDefinitionName)

// 	projectResource := testAccProjectResource(projectName)
// 	return fmt.Sprintf("%s\n%s", projectResource, buildDefinitionResource)
// }

// // Given the name of an AzDO build definition, this will return a function that will check whether
// // or not the definition (1) exists in the state and (2) exist in AzDO and (3) has the correct name
// func testAccCheckBuildDefinitionResourceExists(expectedName string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		buildDef, ok := s.RootModule().Resources["azuredevops_build_definition.build"]
// 		if !ok {
// 			return fmt.Errorf("Did not find a build definition in the TF state")
// 		}

// 		buildDefinition, err := getBuildDefinitionFromResource(buildDef)
// 		if err != nil {
// 			return err
// 		}

// 		if *buildDefinition.Name != expectedName {
// 			return fmt.Errorf("Build Definition has Name=%s, but expected Name=%s", *buildDefinition.Name, expectedName)
// 		}

// 		return nil
// 	}
// }

// // verifies that all build definitions referenced in the state are destroyed. This will be invoked
// // *after* terrafform destroys the resource but *before* the state is wiped clean.
// func testAccBuildDefinitionCheckDestroy(s *terraform.State) error {
// 	for _, resource := range s.RootModule().Resources {
// 		if resource.Type != "azuredevops_build_definition" {
// 			continue
// 		}

// 		// indicates the build definition still exists - this should fail the test
// 		if _, err := getBuildDefinitionFromResource(resource); err == nil {
// 			return fmt.Errorf("Unexpectedly found a build definition that should be deleted")
// 		}
// 	}

// 	return nil
// }

// // given a resource from the state, return a build definition (and error)
// func getBuildDefinitionFromResource(resource *terraform.ResourceState) (*build.BuildDefinition, error) {
// 	buildDefID, err := strconv.Atoi(resource.Primary.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	projectID := resource.Primary.Attributes["project_id"]
// 	clients := testAccProvider.Meta().(*aggregatedClient)
// 	return clients.BuildClient.GetDefinition(clients.ctx, build.GetDefinitionArgs{
// 		Project:      &projectID,
// 		DefinitionId: &buildDefID,
// 	})
// }
