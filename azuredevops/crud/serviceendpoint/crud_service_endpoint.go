package crudserviceendpoint

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/serviceendpoint"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/tfhelper"
)

type flatFunc func(d *schema.ResourceData, serviceEndpoint *serviceendpoint.ServiceEndpoint, projectID *string)
type expandFunc func(d *schema.ResourceData) (*serviceendpoint.ServiceEndpoint, *string)

//GenBaseServiceEndpointResource creates a Resource with the common parts
// that all Service Endpoints require.
func GenBaseServiceEndpointResource(f flatFunc, e expandFunc) *schema.Resource {
	return &schema.Resource{
		Create: genServiceEndpointCreateFunc(f, e),
		Read:   genServiceEndpointReadFunc(f),
		Update: genServiceEndpointUpdateFunc(f, e),
		Delete: genServiceEndpointDeleteFunc(e),
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				// d.Id() here is the last argument passed to the `terraform import RESOURCE_TYPE.RESOURCE_NAME RESOURCE_ID` command
				// Here we use a function to parse the import ID (like the example above) to simplify our logic
				projectID, ServiceID, err := ParseImportedProjectIDAndEndPointName(meta.(*config.AggregatedClient), d.Id())
				if err != nil {
					return nil, fmt.Errorf("Error parsing the service end point from the Terraform resource data: %v", err)
				}
				d.Set("project_id", projectID)
				d.SetId(fmt.Sprintf("%s", ServiceID))
				//tfhelper.HelpFlattenSecret(d,)

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: genBaseSchema(),
	}
}

func genBaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"service_endpoint_name": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

//DoBaseExpansion performs the expansion for the 'base' attributes that are defined in the schema, above
func DoBaseExpansion(d *schema.ResourceData) (*serviceendpoint.ServiceEndpoint, *string) {
	// an "error" is OK here as it is expected in the case that the ID is not set in the resource data
	var serviceEndpointID *uuid.UUID
	parsedID, err := uuid.Parse(d.Id())
	if err == nil {
		serviceEndpointID = &parsedID
	}
	projectID := converter.String(d.Get("project_id").(string))
	serviceEndpoint := &serviceendpoint.ServiceEndpoint{
		Id:    serviceEndpointID,
		Name:  converter.String(d.Get("service_endpoint_name").(string)),
		Owner: converter.String("library"),
	}

	return serviceEndpoint, projectID
}

//DoBaseFlattening performs the flattening for the 'base' attributes that are defined in the schema, above
func DoBaseFlattening(d *schema.ResourceData, serviceEndpoint *serviceendpoint.ServiceEndpoint, projectID *string) {
	d.SetId(serviceEndpoint.Id.String())
	d.Set("service_endpoint_name", *serviceEndpoint.Name)
	d.Set("project_id", projectID)
}

// MakeProtectedSchema create protected schema
func MakeProtectedSchema(r *schema.Resource, keyName, envVarName, description string) {
	r.Schema[keyName] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		DefaultFunc:      schema.EnvDefaultFunc(envVarName, nil),
		Description:      description,
		Sensitive:        true,
		DiffSuppressFunc: tfhelper.DiffFuncSupressSecretChanged,
	}

	secretHashKey, secretHashSchema := tfhelper.GenerateSecreteMemoSchema(keyName)
	r.Schema[secretHashKey] = secretHashSchema
}

// MakeUnprotectedSchema create unprotected schema
func MakeUnprotectedSchema(r *schema.Resource, keyName, envVarName, description string) {
	r.Schema[keyName] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		DefaultFunc: schema.EnvDefaultFunc(envVarName, nil),
		Description: description,
	}
}

// Make the Azure DevOps API call to create the endpoint
func createServiceEndpoint(clients *config.AggregatedClient, endpoint *serviceendpoint.ServiceEndpoint, project *string) (*serviceendpoint.ServiceEndpoint, error) {
	createdServiceEndpoint, err := clients.ServiceEndpointClient.CreateServiceEndpoint(
		clients.Ctx,
		serviceendpoint.CreateServiceEndpointArgs{
			Endpoint: endpoint,
			Project:  project,
		})

	return createdServiceEndpoint, err
}

func deleteServiceEndpoint(clients *config.AggregatedClient, project *string, endPointID *uuid.UUID) error {
	err := clients.ServiceEndpointClient.DeleteServiceEndpoint(
		clients.Ctx,
		serviceendpoint.DeleteServiceEndpointArgs{
			Project:    project,
			EndpointId: endPointID,
		})

	return err
}

func updateServiceEndpoint(clients *config.AggregatedClient, endpoint *serviceendpoint.ServiceEndpoint, project *string) (*serviceendpoint.ServiceEndpoint, error) {
	updatedServiceEndpoint, err := clients.ServiceEndpointClient.UpdateServiceEndpoint(
		clients.Ctx,
		serviceendpoint.UpdateServiceEndpointArgs{
			Endpoint:   endpoint,
			Project:    project,
			EndpointId: endpoint.Id,
		})

	return updatedServiceEndpoint, err
}

func genServiceEndpointCreateFunc(flatFunc flatFunc, expandFunc expandFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		serviceEndpoint, projectID := expandFunc(d)

		createdServiceEndpoint, err := createServiceEndpoint(clients, serviceEndpoint, projectID)
		if err != nil {
			return fmt.Errorf("Error creating service endpoint in Azure DevOps: %+v", err)
		}

		flatFunc(d, createdServiceEndpoint, projectID)
		return nil
	}
}

func genServiceEndpointReadFunc(flatFunc flatFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)

		var serviceEndpointID *uuid.UUID
		parsedServiceEndpointID, err := uuid.Parse(d.Id())
		if err != nil {
			return fmt.Errorf("Error parsing the service endpoint ID from the Terraform resource data: %v", err)
		}
		serviceEndpointID = &parsedServiceEndpointID
		projectID := converter.String(d.Get("project_id").(string))

		serviceEndpoint, err := clients.ServiceEndpointClient.GetServiceEndpointDetails(
			clients.Ctx,
			serviceendpoint.GetServiceEndpointDetailsArgs{
				EndpointId: serviceEndpointID,
				Project:    projectID,
			},
		)
		if err != nil {
			return fmt.Errorf("Error looking up service endpoint given ID (%v) and project ID (%v): %v", serviceEndpointID, projectID, err)
		}

		flatFunc(d, serviceEndpoint, projectID)
		return nil
	}
}

func genServiceEndpointByName(clients *config.AggregatedClient, projectID string, endPointName string) (serviceendpoint.ServiceEndpoint, error) {
	var serviceendpointobj serviceendpoint.ServiceEndpoint

	serviceendlist, err := clients.ServiceEndpointClient.GetServiceEndpointsByNames(
		clients.Ctx,
		serviceendpoint.GetServiceEndpointsByNamesArgs{
			EndpointNames:&[]string{endPointName},
			Project:    &projectID,
		},
	)

	if err != nil {
		return serviceendpointobj,fmt.Errorf("Error looking up service endpoint given name (%v) and project ID (%v): %v", endPointName, projectID, err)
	}

	if len(*serviceendlist) > 0 {
		serviceendpointobj = (*serviceendlist)[0]
		return serviceendpointobj, nil
	}

	return serviceendpointobj, nil
}

func genServiceEndpointUpdateFunc(flatFunc flatFunc, expandFunc expandFunc) schema.UpdateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		serviceEndpoint, projectID := expandFunc(d)

		updatedServiceEndpoint, err := updateServiceEndpoint(clients, serviceEndpoint, projectID)
		if err != nil {
			return fmt.Errorf("Error updating service endpoint in Azure DevOps: %+v", err)
		}

		flatFunc(d, updatedServiceEndpoint, projectID)
		return nil
	}
}

func genServiceEndpointDeleteFunc(expandFunc expandFunc) schema.DeleteFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		serviceEndpoint, projectID := expandFunc(d)

		return deleteServiceEndpoint(clients, projectID, serviceEndpoint.Id)
	}
}


// ParseImportedProjectIDAndEndPointName : Parse the Id (projectId/endpointName) or (projectName/endpointName)
func ParseImportedProjectIDAndEndPointName(clients *config.AggregatedClient, name string) (string, string, error) {
	project, resourceName, err := tfhelper.ParseImportedName(name)
	if err != nil {
		return "", "", err
	}

	// Get the project ID
	currentProject, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
		ProjectId:           &project,
		IncludeCapabilities: converter.Bool(true),
		IncludeHistory:      converter.Bool(false),
	})
	if err != nil {
		return "", "", err
	}

	serviceendpoint, err := genServiceEndpointByName(clients,currentProject.Id.String(), resourceName)
	if err !=nil {

	}

	return currentProject.Id.String(), serviceendpoint.Id.String(), nil
}
