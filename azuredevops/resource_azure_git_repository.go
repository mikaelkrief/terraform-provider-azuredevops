package azuredevops

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
)

func resourceAzureGitRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceAzureGitRepositoryCreate,
		Read:   resourceAzureGitRepositoryRead,
		Update: resourceAzureGitRepositoryUpdate,
		Delete: resourceAzureGitRepositoryDelete,

		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: false,
				Required: true,
			},
			"default_branch": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"is_fork": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"remote_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ssh_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"web_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_remote_urls": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAzureGitRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)
	repo, projectId, err := expandAzureGitRepository(clients, d)

	createdRepo, err := createAzureGitRepository(clients, repo, projectId)
	if err != nil {
		return fmt.Errorf("Error creating project in Azure DevOps: %+v", err)
	}

	flattenAzureGitRepository(clients, d, createdRepo)

	return resourceAzureGitRepositoryRead(d, m)
}

// Make API call to create the repo
func createAzureGitRepository(clients *aggregatedClient, repo *git.GitRepository, projectId *uuid.UUID) (*git.GitRepository, error) {
	args := git.CreateRepositoryArgs{
		GitRepositoryToCreate: &git.GitRepositoryCreateOptions{
			Name: repo.Name,
			Project: &core.TeamProjectReference{
				Id: projectId,
			},
		},
	}
	createdRepository, err := clients.GitReposClient.CreateRepository(clients.ctx, args)

	return createdRepository, err
}

func resourceAzureGitRepositoryRead(d *schema.ResourceData, m interface{}) error {
	repoID := d.Id()
	repoName := d.Get("name").(string)
	projectID := d.Get("project_id").(string)

	clients := m.(*aggregatedClient)
	repo, err := azureGitRepositoryRead(clients, repoID, repoName, projectID)
	if err != nil {
		return fmt.Errorf("Error looking up repository with ID %s and Name %s. Error: %v", repoID, repoName, err)
	}

	flattenAzureGitRepository(clients, d, repo)
	return nil
}

func resourceAzureGitRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceAzureGitRepositoryRead(d, m)
}

func resourceAzureGitRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func deleteAzureGitRepository(clients *aggregatedClient, id string, timeoutSeconds time.Duration) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("Invalid repositoryId UUID: %s", id)
	}

	return clients.GitReposClient.DeleteRepository(clients.ctx, git.DeleteRepositoryArgs{
		RepositoryId: &uuid,
	})
}

// Lookup an Azure Git Repository using the ID, or name if the ID is not set.
func azureGitRepositoryRead(clients *aggregatedClient, repoID string, repoName string, projectID string) (*git.GitRepository, error) {
	identifier := repoID
	if identifier == "" {
		identifier = repoName
	}

	return clients.GitReposClient.GetRepository(clients.ctx, git.GetRepositoryArgs{
		RepositoryId: converter.String(identifier),
		Project:      converter.String(projectID),
	})
}

func flattenAzureGitRepository(clients *aggregatedClient, d *schema.ResourceData, repository *git.GitRepository) {
	d.SetId(repository.Id.String())

	d.Set("name", converter.ToString(repository.Name, ""))
	d.Set("project_id", repository.Project.Id.String())
	d.Set("default_branch", converter.ToString(repository.DefaultBranch, ""))
	d.Set("is_fork", repository.IsFork)
	d.Set("remote_url", converter.ToString(repository.RemoteUrl, ""))
	d.Set("size", repository.Size)
	d.Set("ssh_url", converter.ToString(repository.SshUrl, ""))
	d.Set("url", converter.ToString(repository.Url, ""))
	d.Set("web_url", converter.ToString(repository.WebUrl, ""))
	d.Set("valid_remote_urls", repository.ValidRemoteUrls)
}

// Convert internal Terraform data structure to an AzDO data structure
func expandAzureGitRepository(clients *aggregatedClient, d *schema.ResourceData) (*git.GitRepository, *uuid.UUID, error) {
	// an "error" is OK here as it is expected in the case that the ID is not set in the resource data
	var repoID *uuid.UUID
	parsedID, err := uuid.Parse(d.Id())
	if err == nil {
		repoID = &parsedID
	}

	projectID, err := uuid.Parse(d.Get("project_id").(string))
	if err != nil {
		return nil, nil, err
	}

	repo := &git.GitRepository{
		Id:            repoID,
		Name:          converter.String(d.Get("name").(string)),
		DefaultBranch: converter.String(d.Get("default_branch").(string)),
	}

	return repo, &projectID, nil
}
