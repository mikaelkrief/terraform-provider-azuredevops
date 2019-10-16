package azuredevops

import (
	"fmt"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/graph"
)

func resourceGraphMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceGraphMembershipCreate,
		Read:   resourceGraphMembershipRead,
		Delete: resourceGraphMembershipDelete,

		Schema: map[string]*schema.Schema{
			"container_descriptor": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"subject_descriptor": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceGraphMembershipCreate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)
	graphMembership := expandGraphMembership(d)

	createdGraphMembership, err := clients.GraphClient.AddMembership(clients.ctx, graph.AddMembershipArgs{
		SubjectDescriptor:   graphMembership.MemberDescriptor,
		ContainerDescriptor: graphMembership.ContainerDescriptor,
	})

	if err != nil {
		return err
	}

	flattenGraphMembership(d, createdGraphMembership)
	return nil
}

func resourceGraphMembershipRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*aggregatedClient)
	graphMembership := expandGraphMembership(d)

	currentGraphMembership, err := clients.GraphClient.GetMembership(clients.ctx, graph.GetMembershipArgs{
		SubjectDescriptor:   graphMembership.MemberDescriptor,
		ContainerDescriptor: graphMembership.ContainerDescriptor,
	})

	// an error is only returned in the case that the membership does not exist. It is not an error
	// from the perspective of Terraform, it just means that the state should be updated.
	if err != nil {
		d.SetId("")
		return nil
	}

	flattenGraphMembership(d, currentGraphMembership)
	return nil
}

func resourceGraphMembershipDelete(d *schema.ResourceData, m interface{}) error {
	if d.Id() == "" {
		return nil
	}

	clients := m.(*aggregatedClient)
	graphMembership := expandGraphMembership(d)

	return clients.GraphClient.RemoveMembership(clients.ctx, graph.RemoveMembershipArgs{
		SubjectDescriptor:   graphMembership.MemberDescriptor,
		ContainerDescriptor: graphMembership.ContainerDescriptor,
	})
}

func flattenGraphMembership(d *schema.ResourceData, graphMembership *graph.GraphMembership) {
	d.SetId(graphMembershipToID(graphMembership))
	d.Set("container_descriptor", converter.ToString(graphMembership.ContainerDescriptor, ""))
	d.Set("subject_descriptor", converter.ToString(graphMembership.MemberDescriptor, ""))
}

func expandGraphMembership(d *schema.ResourceData) *graph.GraphMembership {
	graphMembership := graph.GraphMembership{
		ContainerDescriptor: converter.String(d.Get("container_descriptor").(string)),
		MemberDescriptor:    converter.String(d.Get("subject_descriptor").(string)),
	}
	return &graphMembership
}

// There is no natural ID for this resource. However, there is no harm in making our own.
func graphMembershipToID(graphMembership *graph.GraphMembership) string {
	return fmt.Sprintf(
		"%s:%s",
		converter.ToString(graphMembership.ContainerDescriptor, ""),
		converter.ToString(graphMembership.MemberDescriptor, ""))
}
