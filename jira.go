package cli

import (
	"fmt"
	"log"

	"github.com/andygrunwald/go-jira"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

type JiraClientFactory struct{}

func (j *JiraClientFactory) GetClient() (*jira.Client, error) {
	// TODO
	panic("not impl")
}

func NewJiraClientFactory() *JiraClientFactory {
	return &JiraClientFactory{}
}

type IssueLinkTypeMenu struct {
	jiraClientFactory *JiraClientFactory
	issueLinkTypes    []jira.IssueLinkType
	cursor            int
}

func (m *IssueLinkTypeMenu) IssueLinkType() (linkType jira.IssueLinkType, subjectIsInward bool) {
	linkType = m.issueLinkTypes[m.cursor/2]
	subjectIsInward = m.cursor%2 == 0
	return
}

func (m *IssueLinkTypeMenu) Select(subjectIssueId string) error {
	if m.issueLinkTypes == nil {
		client, err := m.jiraClientFactory.GetClient()
		if err != nil {
			return err
		}

		apiEndpoint := fmt.Sprintf("rest/api/2/issueLinkType")
		req, err := client.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			return err
		}

		var issueLinkTypesWrap map[string][]jira.IssueLinkType
		resp, err := client.Do(req, issueLinkTypesWrap)
		LogHttpResponse(resp)
		if err != nil {
			return err
		}
		var prs bool
		m.issueLinkTypes, prs = issueLinkTypesWrap["issueLinkTypes"]
		if !prs {
			return errors.New("Unexpected format, no issueLinkTypes key")
		}
	}

	labels := make([]string, len(m.issueLinkTypes)*2)
	for i, link := range m.issueLinkTypes {
		labels[i*2] = subjectIssueId + " - " + link.Inward + "..."
		labels[i*2+1] = subjectIssueId + " - " + link.Outward + "..."
	}

	p := promptui.Select{
		Label: "Choose a relation",
		Items: labels,
		Size:  10,
	}
	cursor, _, err := p.RunCursorAt(m.cursor, 0)
	if err != nil {
		log.Printf("Error making selection: ", err)
		return err
	}

	m.cursor = cursor
	return nil
}
