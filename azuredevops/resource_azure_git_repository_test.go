package azuredevops

import (
	"context"
	"fmt"
	"github.com/microsoft/terraform-provider-azuredevops/azdosdkmocks"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/stretchr/testify/require"
)

/**
 * Begin unit tests
 */

// verifies that the create operation is considered failed if the initial API
// call fails.
func TestAzureGitRepo_Create_DoesNotSwallowErrorFromFailedCreateCall(t *testing.T) {
}

// verifies that a round-trip flatten/expand sequence will not result in data loss
func TestAzureGitRepo_FlattenExpand_RoundTrip(t *testing.T) {
}

// verifies that the read operation is considered failed if the initial API
// call fails.
func TestAzureGitRepo_Read_DoesNotSwallowErrorFromFailedReadCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &aggregatedClient{
		GitReposClient: reposClient,
		ctx:            context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, resourceAzureGitRepository().Schema, nil)
	resourceData.SetId("an-id")
	resourceData.Set("project_id", "a-project")

	expectedArgs := git.GetRepositoryArgs{RepositoryId: converter.String("an-id"), Project: converter.String("a-project")}
	reposClient.
		EXPECT().
		GetRepository(clients.ctx, expectedArgs).
		Return(nil, fmt.Errorf("GetRepository() Failed")).
		Times(1)

	err := resourceAzureGitRepositoryRead(resourceData, clients)
	require.Contains(t, err.Error(), "GetRepository() Failed")
}

// verifies that the resource ID is used for reads if the ID is set
func TestAzureGitRepo_Read_UsesIdIfSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &aggregatedClient{
		GitReposClient: reposClient,
		ctx:            context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, resourceAzureGitRepository().Schema, nil)
	resourceData.SetId("an-id")
	resourceData.Set("project_id", "a-project")

	expectedArgs := git.GetRepositoryArgs{RepositoryId: converter.String("an-id"), Project: converter.String("a-project")}
	reposClient.
		EXPECT().
		GetRepository(clients.ctx, expectedArgs).
		Return(nil, fmt.Errorf("error")).
		Times(1)

	resourceAzureGitRepositoryRead(resourceData, clients)
}

// verifies that the name is used for reads if the ID is not set
func TestAzureGitRepo_Read_UsesNameIfIdNotSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &aggregatedClient{
		GitReposClient: reposClient,
		ctx:            context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, resourceAzureGitRepository().Schema, nil)
	resourceData.Set("name", "a-name")
	resourceData.Set("project_id", "a-project")

	expectedArgs := git.GetRepositoryArgs{RepositoryId: converter.String("a-name"), Project: converter.String("a-project")}
	reposClient.
		EXPECT().
		GetRepository(clients.ctx, expectedArgs).
		Return(nil, fmt.Errorf("error")).
		Times(1)

	resourceAzureGitRepositoryRead(resourceData, clients)
}

/**
 * Begin acceptance tests
 */

// Verifies that the following sequence of events occurrs without error:
//	(1) TF apply creates resource
//	(2) TF state values are set
//	(3) resource can be queried by ID and has expected name
// 	(4) TF destroy deletes resource
//	(5) resource can no longer be queried by ID
func TestAccAzureGitRepo_Create(t *testing.T) {
}
