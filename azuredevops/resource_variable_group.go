package azuredevops

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"strconv"
	"strings"
)

func resourceVariableGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVariableGroupCreate,
		Read:   resourceVariableGroupRead,
		Update: resourceVariableGroupUpdate,
		Delete: resourceVariableGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		//https://godoc.org/github.com/hashicorp/terraform/helper/schema#Schema
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"allow_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"variables": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"variables_secret": {
				Type:     schema.TypeSet,
				Optional: true,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},

			},
		},
	}
}


func resourceVariableGroupCreate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)
	variableGroup, projectID, err := expandVariableGroup(d, nil)
	if err != nil {
		return fmt.Errorf("Error converting terraform data model to AzDO variable group reference: %+v", err)
	}

	varGroup := &taskagent.VariableGroupParameters{
		Description:  variableGroup.Description,
		Name:         variableGroup.Name,
		Type:         variableGroup.Type,
		Variables:    variableGroup.Variables,
		ProviderData: nil,
	}

	createdVariableGroup, err := clients.TaskAgentClient.AddVariableGroup(clients.ctx, taskagent.AddVariableGroupArgs{
		Group:   varGroup,
		Project: &projectID,
	})

	if err != nil {
		return fmt.Errorf("Error creating variable group in Azure DevOps: %+v", err)
	}

	flattenVariableGroup(d, createdVariableGroup, projectID)
	return nil
}

func resourceVariableGroupUpdate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)

	projectID, variableGroupID, err := GetComputedId(clients, d.Id())
	if err != nil {
		return err
	}

	variableGroup, projectID, err := expandVariableGroup(d, &variableGroupID)
	if err != nil {
		return err
	}

	group := taskagent.VariableGroupParameters{
		Description: variableGroup.Description,
		Name:        variableGroup.Name,
		Type:        variableGroup.Type,
		Variables:   variableGroup.Variables,
	}

	updatedVariableGroup, err := clients.TaskAgentClient.UpdateVariableGroup(m.(*aggregatedClient).ctx, taskagent.UpdateVariableGroupArgs{
		Group:   &group,
		Project: &projectID,
		GroupId: variableGroup.Id,
	})

	if err != nil {
		return err
	}

	flattenVariableGroup(d, updatedVariableGroup, projectID)
	return nil
}

func resourceVariableGroupRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)

	projectID, variableGroupID, err := GetComputedId(clients, d.Id())
	if err != nil {
		return err
	}

	variableGroup, err := clients.TaskAgentClient.GetVariableGroup(clients.ctx, taskagent.GetVariableGroupArgs{
		Project: &projectID,
		GroupId: &variableGroupID,
	})

	if err != nil {
		return err
	}

	flattenVariableGroup(d, variableGroup, projectID)
	return nil
}

func resourceVariableGroupDelete(d *schema.ResourceData, m interface{}) error {
	if d.Id() == "" {
		return nil
	}

	clients := m.(*aggregatedClient)
	projectID, variableGroupID, err := parseIdentifiers(d)
	if err != nil {
		return err
	}

	err = clients.TaskAgentClient.DeleteVariableGroup(m.(*aggregatedClient).ctx, taskagent.DeleteVariableGroupArgs{
		Project: &projectID,
		GroupId: &variableGroupID,
	})

	return err
}

func GetComputedId(clients *aggregatedClient, id string) (string, int, error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, fmt.Errorf("unexpected format of ID (%s), expected projectid/groupId", id)
	}

	projectName := parts[0]
	groupId, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("Error converting getting the variable group id: %+v", err)
	}

	currentProject, err := projectRead(clients, "", projectName)
	if err != nil {
		return "", 0, err
	}

	return currentProject.Id.String(), groupId, nil
}

// Convert internal Terraform data structure to an AzDO data structure
func expandVariableGroup(d *schema.ResourceData, variableGroupId *int) (*taskagent.VariableGroup, string, error) {
	projectID := d.Get("project_id").(string)
	variables := expandVariables(d)

	variableGroup := taskagent.VariableGroup{
		Id:          variableGroupId,
		Name:        converter.String(d.Get("name").(string)),
		Description: converter.String(d.Get("description").(string)),
		Type:        converter.String("Vsts"),
		Variables:   &variables,
	}

	return &variableGroup, projectID, nil
}

func flattenVariableGroup(d *schema.ResourceData, variableGroup *taskagent.VariableGroup, projectID string) {
	d.SetId(fmt.Sprintf("%s/%d", projectID, *variableGroup.Id))

	d.Set("project_id", projectID)
	d.Set("name", *variableGroup.Name)
	d.Set("type", variableGroup.Type)
	d.Set("description", variableGroup.Description)
	d.Set("allow_access", true)
	variables := flattenVariables(*variableGroup.Variables)
	variablesSecret := flattenVariablesSecret(*variableGroup.Variables)
	d.Set("variables", variables)
	d.Set("variables_secret", variablesSecret)

}

func expandVariables(d *schema.ResourceData) map[string]taskagent.VariableValue {
	vars := d.Get("variables").(*schema.Set).List()
	varsSecret := d.Get("variables_secret").(*schema.Set).List()
	output := make(map[string]taskagent.VariableValue, len(vars)+len(varsSecret))

	for _, v := range vars {
		vals := v.(map[string]interface{})

		varName := vals["name"].(string)
		varValue := vals["value"].(string)

		isSecret, _ := strconv.ParseBool("false")
		output[varName] = taskagent.VariableValue{
			Value:    &varValue,
			IsSecret: &isSecret,
		}
	}

	for _, v := range varsSecret {
		vals := v.(map[string]interface{})

		varName := vals["name"].(string)
		//varValue := vals["value"].(string)

		isSecret, _ := strconv.ParseBool("true")
		output[varName] = taskagent.VariableValue{
			//Value:    &varValue,
			IsSecret: &isSecret,
		}
	}

	return output
}

func flattenVariables(input map[string]taskagent.VariableValue) []interface{} {
	results := make([]interface{}, 0)

	for k, v := range input {
		if v.IsSecret == nil {
			result := make(map[string]interface{})
			result["name"] = k
			if v.Value != nil {
				result["value"] = *v.Value
			}
			results = append(results, result)
		}
	}

	return results
}

func flattenVariablesSecret(input map[string]taskagent.VariableValue) []interface{} {
	results := make([]interface{}, 0)

	for k, v := range input {
		if v.IsSecret != nil && *v.IsSecret == true {
			result := make(map[string]interface{})
			result["name"] = k

			results = append(results, result)
		}
	}

	return results
}
