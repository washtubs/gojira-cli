package cli

import "github.com/andygrunwald/go-jira"

type JiraClientFactory struct{}

func (j *JiraClientFactory) GetClient() *jira.Client {
	// TODO
	panic("not impl")
}

func NewJiraClientFactory() *JiraClientFactory {
	return &JiraClientFactory{}
}
