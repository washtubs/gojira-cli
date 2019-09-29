package cli

import (
	"fmt"
	"os"
)

type Options struct {
}

type App struct {
	config *Config

	workbench       *Workbench
	issueSearcher   IssueSearcher
	formatterConfig *FormatterConfig

	jiraClientFactory  *JiraClientFactory
	menuService        *MenuService
	issueFormatter     IssueFormatter
	issueSelector      *IssueSelector
	issueSearchService *IssueSearchService
	executorService    *ExecutorService
	actionBaseService  *ActionBaseService
	workbenchService   WorkbenchService
}

func NewApp() *App {
	app := &App{}

	// Load the config
	app.config = NewConfigLoader().LoadConfig()

	// Create stateful entities
	app.workbench = InitWorkbench()
	app.issueSearcher = InitIssueSearcher()
	app.formatterConfig = &FormatterConfig{}

	// Wire everything up
	app.jiraClientFactory = NewJiraClientFactory()
	app.menuService = NewMenuService(app.config)
	app.issueFormatter = NewIssueFormatter(app.formatterConfig)
	app.issueSelector = &IssueSelector{app.issueFormatter}
	app.issueSearchService = NewIssueSearchService(app.issueSearcher, app.menuService, app.issueSelector)
	app.executorService = NewExecutorService(app.jiraClientFactory)
	app.actionBaseService = NewActionBaseService(app.config, app.menuService, app.issueSearchService)
	app.workbenchService = NewWorkbenchService(app.issueSelector, app.issueSearchService, app.actionBaseService, app.executorService)

	mainMenuActions = MainMenuActions(app.workbenchService, app.menuService, app.workbench)

	mainMenuActionLabels := make([]string, len(mainMenuActions))
	for i, action := range mainMenuActions {
		mainMenuActionLabels[i] = action.label
	}
	app.menuService.RegisterMainMenu(mainMenuActionLabels)
	app.menuService.RegisterFormatterMenu(&FormatterMenu{app.formatterConfig, 0})

	return app
}

var (
	config *Config

	workbench       *Workbench
	issueSearcher   IssueSearcher
	formatterConfig *FormatterConfig

	jiraClientFactory  *JiraClientFactory
	menuService        *MenuService
	issueFormatter     IssueFormatter
	issueSelector      *IssueSelector
	issueSearchService *IssueSearchService
	executorService    *ExecutorService
	actionBaseService  *ActionBaseService
	workbenchService   WorkbenchService

	mainMenuActions []*MenuAction
)

type MenuAction struct {
	action func() error
	label  string
}

func MainMenuActions(svc WorkbenchService, menuService *MenuService, w *Workbench) []*MenuAction {
	return []*MenuAction{
		&MenuAction{
			action: func() error { os.Exit(0); return nil },
			label:  "Quit",
		},
		&MenuAction{
			action: func() error { w.Reset(); return nil },
			label:  "Reset all",
		},
		&MenuAction{
			action: func() error { return nil },
			label:  "Print",
		},
		&MenuAction{
			action: func() error { return svc.AddIssuesInteractive(w) },
			label:  "Seach / Add issues to workbench",
		},
		&MenuAction{
			action: func() error { return svc.AssignInteractive(w) },
			label:  "Queue issues",
		},
		&MenuAction{
			action: func() error { return svc.SelectInteractive(w) },
			label:  "Select issues",
		},
		&MenuAction{
			action: func() error { return svc.FilterInteractive(w) },
			label:  "Remove issues",
		},
		&MenuAction{
			action: func() error { return svc.SelectActionInteractive(w) },
			label:  "Choose action to be assigned",
		},
		&MenuAction{
			action: func() error { return svc.AddActionInteractive(w) },
			label:  "Add action",
		},
		&MenuAction{
			action: func() error { return svc.RemoveActionInteractive(w) },
			label:  "Remove action",
		},
		&MenuAction{
			action: func() error { return svc.EditActionInteractive(w) },
			label:  "Edit action (re-add action, keeping assigned issues)",
		},
		&MenuAction{
			action: func() error { return menuService.SelectIssueFormat() },
			label:  "Change issue format",
		},
		&MenuAction{
			action: func() error { return svc.Execute(w, false) },
			label:  "Execute",
		},
		&MenuAction{
			action: func() error { return svc.Execute(w, true) },
			label:  "Preview",
		},
	}
}

func RunWorkbench() {
	//SetupRpc()

	app := NewApp()

	config = app.config

	workbench = app.workbench
	issueSearcher = app.issueSearcher
	formatterConfig = app.formatterConfig

	menuService = app.menuService
	jiraClientFactory = app.jiraClientFactory
	issueFormatter = app.issueFormatter
	issueSelector = app.issueSelector
	issueSearchService = app.issueSearchService
	actionBaseService = app.actionBaseService
	executorService = app.executorService
	workbenchService = app.workbenchService

	for {

		fmt.Println("-----------------------------")

		// No issues. Must search
		if len(workbench.working)+len(workbench.assigned) == 0 {
			workbenchService.AddIssuesInteractive(workbench)
		}

		fmt.Println(workbench.Format())

		cursor, err := menuService.SelectMain()
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = mainMenuActions[cursor].action()
		if err != nil && !IsCancelError(err) {
			fmt.Println(err)
		}

	}

}
