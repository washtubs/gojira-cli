package cli

import (
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

type Workbench struct {
	working     []jira.Issue
	selection   map[string]bool
	assigned    []IssueAssignment
	actionBases map[int]IssueActionBase
	completed   []IssueAction
	// If assigning while 0, a new action will need to be constructed
	actionId int
	_idIdx   int
}

type IssueAssignment struct {
	actionId int
	issue    jira.Issue
}

func (w *Workbench) Format() string {
	var actionDesc string
	actionBase, prs := w.actionBases[w.actionId]
	if prs {
		// TODO use description
		actionDesc = WrapFormatter(actionBase).Format()
	} else {
		actionDesc = "No action"
	}

	queue := w.Queue()
	queueFmt := ""
	for _, ia := range queue {
		formatter := IssueActionFormatter{ia}
		queueFmt = queueFmt + formatter.Format() + "\n"
	}

	return fmt.Sprintf(`
Current action: %s
Issues: %d
Selected: %d
Queued: %d
Completed: %d
Actions: %d
Queue: %d
%s
`, actionDesc, len(w.working), len(w.selection), len(w.assigned), len(w.completed), len(w.actionBases), len(queue), queueFmt)

}

func (w *Workbench) dupeActionChecker() map[string]bool {
	checker := make(map[string]bool)
	for _, action := range w.actionBases {
		checker[canonicalAction(action)] = true
	}
	return checker
}

func (w *Workbench) dupeAssignedChecker() map[string]bool {
	checker := make(map[string]bool)
	for _, assigned := range w.assigned {
		action := w.actionBases[assigned.actionId]
		checker[assigned.issue.ID+" "+canonicalAction(action)] = true
	}
	return checker
}

func (w *Workbench) AddActionBase(actionBase IssueActionBase) (int, error) {
	checker := w.dupeActionChecker()
	if _, prs := checker[canonicalAction(actionBase)]; prs {
		return 0, errors.New("Action already exists")
	}
	w._idIdx = w._idIdx + 1
	w.actionBases[w._idIdx] = actionBase
	return w._idIdx, nil
}

// Deletes the actionBase
// Invalidates assigned
func (w *Workbench) RemoveActionBase(actionId int) {
	delete(w.actionBases, actionId)
	assigned := make([]IssueAssignment, 0)
	for _, v := range w.assigned {
		if actionId != v.actionId {
			assigned = append(assigned, v)
		}
	}
	w.assigned = assigned
}

func (w *Workbench) Selected() []jira.Issue {
	selected := make([]jira.Issue, 0, len(w.selection))
	for _, issue := range w.working {
		if _, prs := w.selection[issue.ID]; prs {
			selected = append(selected, issue)
		}
	}
	return selected
}

// Assign a particular jira issue from working
func (w *Workbench) Assign(issue jira.Issue) {

	checker := w.dupeAssignedChecker()
	action := w.actionBases[w.actionId]

	if _, prs := checker[issue.ID+" "+canonicalAction(action)]; prs {
		return
	}
	w.assigned = append(w.assigned, IssueAssignment{w.actionId, issue})
}

func (w *Workbench) AssignSelected() {
	if w.actionId == 0 {
		panic("No action to assign to")
	}
	for _, issue := range w.Selected() {
		w.Assign(issue)
	}
}

// Removes selected issues from workings
// The entire selection is invalidated / cleared
func (w *Workbench) RemoveSelected() {
	newWorking := make([]jira.Issue, 0, len(w.working)-len(w.selection))

	for _, issue := range w.working {
		if _, prs := w.selection[issue.ID]; !prs {
			newWorking = append(newWorking, issue)
		}
	}

	w.working = newWorking

	w.ClearSelection()

}

func (w *Workbench) Select(workingIdxs []int) {
	w.ClearSelection()
	for _, idx := range workingIdxs {
		w.selection[w.working[idx].ID] = true
	}
}

func (w *Workbench) ClearSelection() {
	w.selection = make(map[string]bool)
}

func (w *Workbench) AddIssues(issues []jira.Issue) {
	w.working = mergeJiraIssues(issues, w.working)
}

func (w *Workbench) Queue() []IssueAction {
	queue := make([]IssueAction, len(w.assigned))
	for i, assigned := range w.assigned {
		queue[i] = IssueAction{assigned.issue, w.actionBases[assigned.actionId]}
	}
	return queue
}

// Clears queue at every slot where there isn't an error
func (w *Workbench) ClearQueue(errs []error) {
	assigned := make([]IssueAssignment, 0)
	for i, err := range errs {
		if err != nil {
			assigned = append(assigned, w.assigned[i])
		}
	}
	w.assigned = assigned

}

func (w *Workbench) Reset() {
	w.working = make([]jira.Issue, 0)
	w.selection = make(map[string]bool)
	w.assigned = make([]IssueAssignment, 0)
	w.actionBases = make(map[int]IssueActionBase)
	w.completed = make([]IssueAction, 0)
	w.actionId = 0
	w._idIdx = 0
}

func InitWorkbench() *Workbench {
	return &Workbench{
		working:     make([]jira.Issue, 0),
		selection:   make(map[string]bool),
		assigned:    make([]IssueAssignment, 0),
		actionBases: make(map[int]IssueActionBase),
		completed:   make([]IssueAction, 0),
		actionId:    0,
		_idIdx:      0,
	}
}
