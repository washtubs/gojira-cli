package cli

// Handles all interactive interaction w/ the Workbench
type WorkbenchService interface {

	// Interactively add to the issue collection from a Jira result set
	AddIssuesInteractive(w *Workbench) error

	// Removes selected issues,
	// making a temporary issue selection if needed
	FilterInteractive(w *Workbench) error

	// Select issues from working
	SelectInteractive(w *Workbench) error

	// Assigns selected issues to the selected action,
	// making a temporary issue selection if needed
	// and selecting an action temporarily if needed
	AssignInteractive(w *Workbench) error

	// Sets the actionId from the workspace to what the user chooses
	SelectActionInteractive(w *Workbench) error

	// Remove actionbases interactively
	RemoveActionInteractive(w *Workbench) error

	// Replace actionbases interactively
	EditActionInteractive(w *Workbench) error

	// Add an action base
	AddActionInteractive(w *Workbench) error

	Execute(w *Workbench, dryRun bool) error
}

type defaultWorkbenchService struct {
	issueSelector      *IssueSelector
	issueSearchService *IssueSearchService
	actionBaseService  *ActionBaseService
	executorService    *ExecutorService
}

// Interactively remove issues from working
func (s *defaultWorkbenchService) FilterInteractive(w *Workbench) error {
	if len(w.selection) == 0 {
		err := s.doSelect(w, "Select issues to remove from the workbench")
		if err != nil {
			return err
		}

	}

	w.RemoveSelected()

	return nil
}

// Interactively add to the issue collection from a Jira result set
func (s *defaultWorkbenchService) AddIssuesInteractive(w *Workbench) error {
	issues, err := s.issueSearchService.SearchInteractive(SelectOptions{
		Prompt: "Select issues to add to the workbench",
	})
	if err != nil {
		return err
	}

	w.AddIssues(issues)
	return nil
}

func (s *defaultWorkbenchService) doSelect(w *Workbench, prompt string) error {
	selected, canceled, err := s.issueSelector.SelectSlc(w.working, SelectOptions{
		Prompt: prompt,
	})
	if err != nil {
		return err
	}
	if canceled {
		return CancelError()
	}
	w.Select(selected)
	return nil
}

func (s *defaultWorkbenchService) SelectInteractive(w *Workbench) error {
	return s.doSelect(w, "Make a selection")
}

func (s *defaultWorkbenchService) AssignInteractive(w *Workbench) error {

	// Make sure we have an action to assign to
	clearAction := false
	if w.actionId == 0 {
		clearAction = true
		err := s.SelectActionInteractive(w)
		if err != nil {
			return err
		}
	}

	// If there is no selection create one to be cleared out at the end
	clearSelection := false
	if len(w.selection) == 0 {
		clearSelection = true

		var actionDesc string
		actionBase, prs := w.actionBases[w.actionId]
		if prs {
			// TODO use description
			actionDesc = WrapFormatter(actionBase).Format()
		} else {
			actionDesc = "an action"
		}

		err := s.doSelect(w, "Select issues to assign to "+actionDesc)
		if err != nil {
			return err
		}
	}

	w.AssignSelected()

	if clearAction {
		w.actionId = 0
	}

	if clearSelection {
		w.ClearSelection()
	}

	return nil
}

// Sets the actionId from the workspace to what the user chooses
func (s *defaultWorkbenchService) SelectActionInteractive(w *Workbench) error {
	if len(w.actionBases) == 0 {
		actionBase, err := s.actionBaseService.BuildAction()
		if err != nil {
			return err
		}

		actionId, err := w.AddActionBase(actionBase)
		if err != nil {
			return err
		}

		w.actionId = actionId
		return nil
	}

	actionId, err := s.actionBaseService.SelectAction(w.actionBases)
	if err != nil {
		return err
	}

	w.actionId = actionId
	return nil
}

// Remove actionbases interactively
func (s *defaultWorkbenchService) RemoveActionInteractive(w *Workbench) error {
	actionId, err := s.actionBaseService.SelectAction(w.actionBases)
	if err != nil {
		return err
	}

	w.RemoveActionBase(actionId)
	return nil
}

func (s *defaultWorkbenchService) EditActionInteractive(w *Workbench) error {
	panic("not impl")
}

// Construct and Add an action base, also sets the actionId
func (s *defaultWorkbenchService) AddActionInteractive(w *Workbench) error {
	actionBase, err := s.actionBaseService.BuildAction()
	if err != nil {
		return err
	}

	actionId, err := w.AddActionBase(actionBase)
	if err != nil {
		return err
	}

	w.actionId = actionId
	return nil
}

func (s *defaultWorkbenchService) Execute(w *Workbench, dryRun bool) error {
	errs := s.executorService.Execute(w.Queue(), dryRun)
	if !dryRun {
		w.ClearQueue(errs)
	}
	return nil
}

func NewWorkbenchService(
	issueSelector *IssueSelector,
	issueSearchService *IssueSearchService,
	actionBaseService *ActionBaseService,
	executorService *ExecutorService,
) WorkbenchService {
	return &defaultWorkbenchService{
		issueSelector,
		issueSearchService,
		actionBaseService,
		executorService,
	}
}
