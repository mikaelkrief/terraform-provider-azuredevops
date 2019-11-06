package azuredevops

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"strconv"
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
						"is_secret": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func resourceVariableGroupCreate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)
	variableGroup, projectID, err := expandVariableGroup(d)
	if err != nil {
		return fmt.Errorf("Error converting terraform data model to AzDO variable group reference: %+v", err)
	}

	group := &taskagent.VariableGroupParameters{
		Description: variableGroup.Description,
		Name:        variableGroup.Name,
		Type:        variableGroup.Type,
		Variables:   variableGroup.Variables,
		ProviderData: nil,
	}


	createdVariableGroup, err := createVariableGroup(clients, group, projectID)
	if err != nil {
		return fmt.Errorf("Error creating variable group in Azure DevOps: %+v", err)
	}

	flattenVariableGroup(d, createdVariableGroup, projectID)
	return nil
}

func createVariableGroup(clients *aggregatedClient, variableGroup *taskagent.VariableGroupParameters, project string) (*taskagent.VariableGroup, error) {


	createdVariableGroup, err := clients.TaskAgentClient.AddVariableGroup(clients.ctx, taskagent.AddVariableGroupArgs{
		Group:   variableGroup,
		Project: &project,
	})

	return createdVariableGroup, err
}

// Convert internal Terraform data structure to an AzDO data structure
func expandVariableGroup(d *schema.ResourceData) (*taskagent.VariableGroup, string, error) {
	projectID := d.Get("project_id").(string)
	variables := expandVariables(d)

	// Look for the ID. This may not exist if we are within the context of a "create" operation,
	// so it is OK if it is missing.

	variableGroupID, err := strconv.Atoi(d.Id())
	var variableGroupReference *int
	if err == nil {
		variableGroupReference = &variableGroupID
	} else {
		variableGroupReference = nil
	}

	variableGroup := taskagent.VariableGroup{
		Id:          variableGroupReference,
		Name:        converter.String(d.Get("name").(string)),
		Description: converter.String(d.Get("description").(string)),
		Type:        converter.String("Vsts"),
		Variables:   &variables,
	}

	return &variableGroup, projectID, nil
}



func flattenVariableGroup(d *schema.ResourceData, variableGroup *taskagent.VariableGroup, projectID string) {
	d.SetId(strconv.Itoa(*variableGroup.Id))

	d.Set("project_id", projectID)
	d.Set("name", *variableGroup.Name)
	variables := flattenVariables(*variableGroup.Variables)
	d.Set("variables", variables)
	d.Set("type", variableGroup.Type)

}



/*
func expandVariables(d *schema.ResourceData) map[string]taskagent.VariableValue {
	if vars := d.Get("variables").(*schema.Set).List(); vars.Len() > 0 {
		_variables := make(map[string]taskagent.VariableValue, vars.Len())
		for _, vrVar := range vars.List() {
			vrVar := vrVar.(map[string]interface{})

			name := vrVar["name"].(string)
			value := vrVar["value"].(string)

			isSecret := vrVar["is_secret"].(bool)
			variable := taskagent.VariableValue{
				Value:    &value,
				IsSecret: &isSecret,
			}
			valueMap := variable

			_variables[name] = valueMap
		}
		return _variables
	}
	return nil
}
*/

func expandVariables(d *schema.ResourceData) map[string]taskagent.VariableValue {
	input := d.Get("variables").(*schema.Set).List()
	output := make(map[string]taskagent.VariableValue, len(input))

	for _, v := range input {
		vals := v.(map[string]interface{})

		varName := vals["name"].(string)
		varValue := vals["value"].(string)
		varIsSecret := vals["is_secret"].(bool)

		output[varName] = taskagent.VariableValue{
			Value: &varValue,
			IsSecret: &varIsSecret,
		}
	}

	return output
}


/*
func flattenVariables(variables map[string]taskagent.VariableValue) map[string]interface{} {

	var out = make(map[string]interface{})

	for i, v := range variables {
		m := make(map[string]interface{})
		m["name"] = i
		m["value"] = v.Value
		m["is_secret"] = v.IsSecret

		out[i] = m
	}


	return out

}*/

func flattenVariables(input map[string]taskagent.VariableValue) []interface{} {
	results := make([]interface{}, 0)

	for k, v := range input {
		result := make(map[string]interface{})
		result["name"] = k
		if v.IsSecret != nil {
			result["is_secret"] = *v.IsSecret
		}
		if v.Value != nil {
			result["value"] = *v.Value
		}
		results = append(results, result)
	}

	return results
}


func resourceVariableGroupUpdate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)
	variableGroup, projectID, err := expandVariableGroup(d)
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
	projectID, variableGroupID, err := parseIdentifiers(d)

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
