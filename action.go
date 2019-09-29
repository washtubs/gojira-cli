package cli

import "github.com/andygrunwald/go-jira"

type IssueAction struct {
	issue  jira.Issue
	action IssueActionBase
}

//type AddLinkAction struct {
//outwardId string
//}

//type NavigateAction struct {
//}
