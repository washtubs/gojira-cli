package cli

import (
	"log"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

type IssueSearchService struct {
	searcher    IssueSearcher
	menuService *MenuService
	selector    *IssueSelector
}

// Handles all interaction with the stateful searcher, allowing menu usage as well
func (s *IssueSearchService) SearchInteractive(opts SelectOptions, useQueryRunner bool) ([]jira.Issue, error) {

	log.Printf("Searching %+v", opts)
	jqlKey, err := s.menuService.SelectJQL()
	if err != nil {
		return nil, err
	}

	jql, prs := config.JQLs[jqlKey]
	if !prs {
		log.Fatalf("JQL key %s not present in map %s", jqlKey, config.JQLs)
	}

	jql = strings.ReplaceAll(jql, "\n", " ")

	s.searcher.SetSearchQuery(jql)

	log.Printf("Executing search, %+v", opts)

	var (
		idxs       []int
		cancelled  bool
		interactor SearchInteractor
	)

	// TODO allow pagination
	issuesChan, interactor := s.searcher.SearchAsync()

	if useQueryRunner {
		idxs, cancelled, err = executeQueryRunner(s.searcher, s.selector.formatter, issuesChan, interactor, opts)
	} else {
		idxs, cancelled, err = s.selector.Select(issuesChan, interactor, opts)
	}
	if err != nil {
		log.Printf("Error encountered during execution of selector query_runner=%v: %s", useQueryRunner, err.Error())
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

type IssueSearchMenu struct {
	workbench           *Workbench
	issueSelector       *IssueSelector
	issueSearchService  *IssueSearchService
	workbenchElseGlobal bool
	cursor              int
}

func (m *IssueSearchMenu) Select() error {
	p := promptui.Select{
		Label: "Choose an issue source",
		Items: []string{"Workbench", "Global search"},
		Size:  10,
	}
	cursor, _, err := p.RunCursorAt(m.cursor, 0)
	if err != nil {
		log.Printf("Error making selection: %s", err)
		return err
	}

	m.cursor = cursor
	m.workbenchElseGlobal = m.cursor == 0
	return nil

}

func (m *IssueSearchMenu) Search(prompt string) (jira.Issue, error) {
	opts := SelectOptions{
		Prompt: prompt,
		One:    true,
	}
	var (
		selected []jira.Issue
		err      error
	)
	if m.workbenchElseGlobal {
		var (
			selectedIdxs []int
			canceled     bool
		)
		selectedIdxs, canceled, err = m.issueSelector.SelectSlc(m.workbench.working, opts)
		if canceled {
			err = CancelError()
		}
		selected = make([]jira.Issue, len(selectedIdxs))
		for i, idx := range selectedIdxs {
			selected[i] = m.workbench.working[idx]
		}
	} else {
		selected, err = m.issueSearchService.SearchInteractive(opts, true)
	}

	if len(selected) != 1 {
		return jira.Issue{}, errors.New("Expected exactly 1 issue")
	}

	return selected[0], err
}
