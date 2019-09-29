package cli

import (
	"fmt"
	"time"

	"github.com/andygrunwald/go-jira"
)

type ExecutorService struct {
	jiraClientFactory *JiraClientFactory
	rateLimiter       chan time.Time
}

// Executes actions indicated which ones failed by index
func (e *ExecutorService) Execute(actions []IssueAction, dryRun bool) []error {
	errs := make([]error, len(actions))

	var client *jira.Client
	if !dryRun {
		client = e.jiraClientFactory.GetClient()
	}

	for i, issueAction := range actions {
		formatter := IssueActionFormatter{issueAction}
		<-e.rateLimiter
		fmt.Println("Executing " + formatter.Format())
		if !dryRun {
			errs[i] = issueAction.action.Execute(issueAction.issue, client)
		}
		if errs[i] != nil {
			fmt.Println("Error occurred during execution: " + errs[i].Error())
		}
	}

	return errs
}

func NewExecutorService(
	jiraClientFactory *JiraClientFactory,
) *ExecutorService {
	rateLimiter := NewRateLimiter(time.Second/4, 5)
	return &ExecutorService{
		jiraClientFactory,
		rateLimiter,
	}
}