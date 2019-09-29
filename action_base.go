package cli

import (
	"github.com/andygrunwald/go-jira"
)

type IssueActionBase interface {
	Action
	Execute(issue jira.Issue, client *jira.Client) error
	// Creates a copy of the action base after building it interactively
	Build(svc *ActionBaseService) (IssueActionBase, error)
	BuildParams(params []string) (IssueActionBase, error)
	ToParams() []string
	IsBuilt() bool
}

type Action interface {
	Key() string
	Description() string
	Template() string
}

type ActionType struct {
	key         string
	description string
	template    string
}

func (a ActionType) Key() string         { return a.key }
func (a ActionType) Description() string { return a.description }
func (a ActionType) Template() string    { return a.template }
func (a ActionType) Format() string      { return a.description }

type BaseAction struct {
	built bool
}

func (a BaseAction) IsBuilt() bool { return a.built }

// Start action definitions

type AddCommentAction struct {
	ActionType
	BaseAction
	Comment string
}

func (a AddCommentAction) Execute(issue jira.Issue, client *jira.Client) error {
	panic("not impl")
}
func (a AddCommentAction) Build(svc *ActionBaseService) (IssueActionBase, error) {
	panic("not impl")
}
func (a AddCommentAction) BuildParams(params []string) (IssueActionBase, error) {
	panic("not impl")
}
func (a AddCommentAction) ToParams() []string { return []string{a.Comment} }

type AddLabelAction struct {
	ActionType
	BaseAction
	Label Label
}

func (a AddLabelAction) Execute(issue jira.Issue, client *jira.Client) error {
	panic("not impl")
}

func (a AddLabelAction) Build(svc *ActionBaseService) (IssueActionBase, error) {
	formatters := make([]Formatter, len(svc.config.LabelsAllowed))
	for i, v := range svc.config.LabelsAllowed {
		formatters[i] = v
	}
	idxs, cancelled, err := FzfSelect(formatters, SelectOptions{
		Prompt: "Please select a label to add",
		One:    true,
	}, 0)
	if err != nil {
		return nil, err
	}
	if cancelled {
		return nil, CancelError()
	}
	if len(idxs) != 1 {
		panic("expected exactly one")
	}
	return AddLabelAction{
		a.ActionType,
		BaseAction{true},
		svc.config.LabelsAllowed[idxs[0]],
	}, nil
}

func (a AddLabelAction) BuildParams(params []string) (IssueActionBase, error) {
	panic("not impl")
}

func (a AddLabelAction) ToParams() []string { return []string{string(a.Label)} }

// End action definitions

var actions = []IssueActionBase{
	AddCommentAction{
		ActionType: ActionType{"addComment", "Add comment", "Add comment to _ISSUE"},
	},
	AddLabelAction{
		ActionType: ActionType{"addLabel", "Add label", "Add label '{{.Label}}' to _ISSUE"},
	},
}

type ActionBaseService struct {
	config             *Config
	menuService        *MenuService
	issueSearchService *IssueSearchService
}

func (s *ActionBaseService) SelectAction(actionBases map[int]IssueActionBase) (int, error) {
	idxToId := make(map[int]int)
	actionBasesSlc := make([]Formatter, len(actionBases))
	idx := 0
	for id, action := range actionBases {
		idxToId[idx] = id
		actionBasesSlc[idx] = WrapFormatter(action)
		idx = idx + 1
	}

	idxs, cancelled, err := FzfSelect(actionBasesSlc, SelectOptions{
		Prompt: "Please select an action",
		One:    true,
	}, 0)
	if err != nil {
		return 0, err
	}
	if cancelled {
		return 0, CancelError()
	}
	if len(idxs) != 1 {
		panic("Expected exactly one")
	}
	return idxToId[idxs[0]], nil
}

func (s *ActionBaseService) BuildAction() (IssueActionBase, error) {
	actionTypeItems := make([]Formatter, len(actions))
	for i, v := range actions {
		actionTypeItems[i] = WrapFormatter(v)
	}
	idxs, cancelled, err := FzfSelect(actionTypeItems, SelectOptions{
		Prompt: "Please select an action type",
		One:    true,
	}, 0)
	if err != nil {
		return nil, err
	}
	if cancelled {
		return nil, CancelError()
	}
	if len(idxs) != 1 {
		panic("Expected exactly one")
	}
	actionBase := actions[idxs[0]]

	built, err := actionBase.Build(s)
	if err != nil {
		return nil, err
	}
	if !built.IsBuilt() {
		panic("Action base is not built")
	}
	return built, nil
}

func NewActionBaseService(
	config *Config,
	menuService *MenuService,
	issueSearchService *IssueSearchService,
) *ActionBaseService {
	return &ActionBaseService{
		config,
		menuService,
		issueSearchService,
	}
}
