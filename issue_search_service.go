package cli

import (
	"log"

	"github.com/andygrunwald/go-jira"
)

type IssueSearchService struct {
	searcher    IssueSearcher
	menuService *MenuService
	selector    *IssueSelector
}

// Handles all interaction with the stateful searcher, allowing menu usage as well
func (s *IssueSearchService) SearchInteractive(opts SelectOptions) ([]jira.Issue, error) {

	jqlKey, err := s.menuService.SelectJQL()
	if err != nil {
		return nil, err
	}

	jql, prs := config.JQLs[jqlKey]
	if !prs {
		log.Fatalf("JQL key %s not present in map %s", jqlKey, config.JQLs)
	}

	s.searcher.SetSearchQuery(jql)

	// TODO allow pagination
	issuesChan, interactor := s.searcher.SearchAsync()

	idxs, cancelled, err := s.selector.Select(issuesChan, interactor, opts)
	if err != nil {
		return nil, err
	}
	if cancelled {
		// TODO: give a menu to change pages
		return nil, CancelError()
	}

	issues := interactor.Loaded()
	selected := make([]jira.Issue, 0, len(issues))
	for _, idx := range idxs {
		selected = append(selected, issues[idx])
	}
	return selected, nil
}

func NewIssueSearchService(
	searcher IssueSearcher,
	menuService *MenuService,
	selector *IssueSelector,
) *IssueSearchService {
	return &IssueSearchService{
		searcher,
		menuService,
		selector,
	}
}
