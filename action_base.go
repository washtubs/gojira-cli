package cli

import (
	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
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

// Add comment

type AddCommentAction struct {
	ActionType
	BaseAction
	Comment string
}

func (a AddCommentAction) Execute(issue jira.Issue, client *jira.Client) error {
	_, resp, err := client.Issue.AddComment(issue.ID, &jira.Comment{Body: a.Comment})
	LogHttpResponse(resp)
	return err
}

func (a AddCommentAction) Build(svc *ActionBaseService) (IssueActionBase, error) {
	a.Comment = svc.menuService.Comment("Leave a comment")
	if a.Comment == "" {
		return nil, errors.New("Comment can not be empty")
	}
	return AddCommentAction{
		a.ActionType,
		BaseAction{true},
		a.Comment,
	}, nil
}

func (a AddCommentAction) BuildParams(params []string) (IssueActionBase, error) {
	panic("not impl")
}

func (a AddCommentAction) ToParams() []string { return []string{a.Comment} }

// Add label

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

// Relate one

type RelateOneAction struct {
	ActionType
	BaseAction
	SubjectIssue    jira.Issue
	IssueLinkType   jira.IssueLinkType
	SubjectIsInward bool
	Comment         string
}

func (a RelateOneAction) Execute(objectIssue jira.Issue, client *jira.Client) error {
	var (
		outwardIssue jira.Issue
		inwardIssue  jira.Issue
	)
	if a.SubjectIsInward {
		inwardIssue = a.SubjectIssue
		outwardIssue = objectIssue
	} else {
		inwardIssue = objectIssue
		outwardIssue = a.SubjectIssue
	}
	resp, err := client.Issue.AddLink(&jira.IssueLink{
		ID:           "",
		Self:         "",
		Type:         a.IssueLinkType,
		OutwardIssue: &outwardIssue,
		InwardIssue:  &inwardIssue,
		Comment:      &jira.Comment{Body: a.Comment},
	})
	LogHttpResponse(resp)
	return err
}

func (a RelateOneAction) Build(svc *ActionBaseService) (IssueActionBase, error) {
	err := svc.menuService.issueSearchMenu.Select()
	if err != nil {
		return nil, err
	}

	subjectIssue, err := svc.menuService.issueSearchMenu.Search("Select an issue as a starting point")
	if err != nil {
		return nil, err
	}

	err = svc.menuService.issueLinkTypeMenu.Select(subjectIssue.ID)
	if err != nil {
		return nil, err
	}

	comment := svc.menuService.Comment("Leave a comment (optional)")

	issueLinkType, subjectIsInward := svc.menuService.issueLinkTypeMenu.IssueLinkType()

	return RelateOneAction{
		a.ActionType,
		BaseAction{true},
		subjectIssue,
		issueLinkType,
		subjectIsInward,
		comment,
	}, nil
}

func (a RelateOneAction) BuildParams(params []string) (IssueActionBase, error) {
	panic("not impl")
}

func (a RelateOneAction) ToParams() []string {
	subjectIsInward := "false"
	if a.SubjectIsInward {
		subjectIsInward = "true"
	}
	return []string{a.SubjectIssue.ID, a.IssueLinkType.Name, subjectIsInward, a.Comment}
}

// Navigate action

type NavigateAction struct {
	ActionType
	BaseAction
}

func (a NavigateAction) Execute(issue jira.Issue, client *jira.Client) error {
	url := client.GetBaseURL()
	openbrowser(url.String() + "/" + issue.ID)
	return nil
}

func (a NavigateAction) Build(svc *ActionBaseService) (IssueActionBase, error) {
	return NavigateAction{
		a.ActionType,
		BaseAction{true},
	}, nil
}

func (a NavigateAction) BuildParams(params []string) (IssueActionBase, error) {
	return NavigateAction{
		a.ActionType,
		BaseAction{true},
	}, nil
}

func (a NavigateAction) ToParams() []string { return []string{} }

// End action definitions

var actions = []IssueActionBase{
	AddCommentAction{
		ActionType: ActionType{"addComment", "Add comment", "Add comment to _ISSUE"},
	},
	AddLabelAction{
		ActionType: ActionType{"addLabel", "Add label", "Add label '{{.Label}}' to _ISSUE"},
	},
	RelateOneAction{
		ActionType: ActionType{"relateOne", "Link issue", "Add link: {{.SubjectIssue.ID}}{{if .SubjectIsInward}} {{.IssueLinkType.Inward}} {{else}} {{.IssueLinkType.Outward}} {{end}}_ISSUE"},
	},
	NavigateAction{
		ActionType: ActionType{"navigate", "Open in browser", "Open _ISSUE in browser"},
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
