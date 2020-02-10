package client

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/mvisonneau/tfcw/lib/schemas"
	log "github.com/sirupsen/logrus"
)

type TFEVariables map[tfe.CategoryType]map[string]*tfe.Variable

func (c *Client) getWorkspace(organization, workspace string) (*tfe.Workspace, error) {
	return c.TFE.Workspaces.Read(c.Context, organization, workspace)
}

func (c *Client) setVariable(w *tfe.Workspace, v *schemas.Variable, e TFEVariables) (*tfe.Variable, error) {
	if v.Sensitive == nil {
		v.Sensitive = tfe.Bool(true)
	}

	if v.HCL == nil {
		v.HCL = tfe.Bool(false)
	}

	if existingVariable, ok := e[getCategoryType(v.Kind)][v.Name]; ok {
		updatedVariable, err := c.TFE.Variables.Update(c.Context, w.ID, existingVariable.ID, tfe.VariableUpdateOptions{
			Key:       &v.Name,
			Value:     v.Value,
			Sensitive: v.Sensitive,
			HCL:       v.HCL,
		})

		// In case we cannot update the fields, we delete the variable and recreate it
		if err != nil {
			log.Debugf("could not update variable id %s, attempting to recreate it.", existingVariable.ID)
			err = c.TFE.Variables.Delete(c.Context, w.ID, existingVariable.ID)
			if err != nil {
				return nil, err
			}
		} else {
			return updatedVariable, nil
		}
	}

	return c.TFE.Variables.Create(c.Context, w.ID, tfe.VariableCreateOptions{
		Key:       &v.Name,
		Value:     v.Value,
		Category:  tfe.Category(getCategoryType(v.Kind)),
		Sensitive: v.Sensitive,
		HCL:       v.HCL,
	})
}

func (c *Client) listVariables(w *tfe.Workspace) (TFEVariables, error) {
	variables := TFEVariables{}

	listOptions := tfe.VariableListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 1,
			PageSize:   20,
		},
	}

	for {
		list, err := c.TFE.Variables.List(c.Context, w.ID, listOptions)
		if err != nil {
			return variables, fmt.Errorf("Unable to list variables from the Terraform Cloud API : %v", err.Error())
		}

		for _, v := range list.Items {
			if _, ok := variables[v.Category]; !ok {
				variables[v.Category] = map[string]*tfe.Variable{}
			}
			variables[v.Category][v.Key] = v
		}

		if list.Pagination.CurrentPage >= list.Pagination.TotalPages {
			break
		}

		listOptions.PageNumber = list.Pagination.NextPage
	}
	return variables, nil
}

func getCategoryType(kind schemas.VariableKind) tfe.CategoryType {
	switch kind {
	case schemas.VariableKindEnvironment:
		return tfe.CategoryEnv
	case schemas.VariableKindTerraform:
		return tfe.CategoryTerraform
	}

	return tfe.CategoryType("")
}
