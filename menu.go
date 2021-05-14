package cli

import (
	"log"

	"github.com/manifoldco/promptui"
)

type StaticMenu struct {
	prompt  string
	entries []string
	cursor  int
}

func (m *StaticMenu) Select() error {
	p := promptui.Select{
		Label: m.prompt,
		Items: m.entries,
		Size:  10,
	}
	var err error
	m.cursor, _, err = p.RunCursorAt(m.cursor, 0)
	if err != nil {
		log.Printf("Error making selection: %s", err)
	}
	return err
}

type FzfMenu struct {
	prompt  string
	entries []string
	cursor  int
}

func (m *FzfMenu) Select() error {
	var err error
	formatters := make([]Formatter, 0, len(m.entries))
	for _, v := range m.entries {
		formatters = append(formatters, StringFormatter(v))
	}

	idxs, cancelled, err := FzfSelect(formatters, SelectOptions{
		Prompt: "Please select an action type",
		One:    true,
	}, 0)
	if err != nil {
		return err
	}
	if cancelled {
		return CancelError()
	}
	if len(idxs) != 1 {
		panic("Expected exactly one")
	}
	m.cursor = idxs[0]

	return err
}

// Basically mediates access to the simple promts using promptui
type MenuService struct {
	config            *Config
	jqlMenu           *FzfMenu
	mainMenu          *StaticMenu
	advancedMenu      *StaticMenu
	formatterMenu     *FormatterMenu
	issueSearchMenu   *IssueSearchMenu
	issueLinkTypeMenu *IssueLinkTypeMenu
	userFavoritesMenu *UsersFavoritesMenu
}

func (s *MenuService) Comment(prompt string) string {
	p := promptui.Prompt{
		Label: prompt,
	}
	comment, err := p.Run()
	if err != nil {
		log.Println("Got error: " + err.Error())
	}

	return comment
}

// Interactively select a JQL key
func (s *MenuService) SelectJQL() (string, error) {
	err := s.jqlMenu.Select()
	return keysFromMap(s.config.JQLs)[s.jqlMenu.cursor], err
}

func (s *MenuService) SelectMain() (int, error) {
	s.mainMenu.cursor = 4
	err := s.mainMenu.Select()
	return s.mainMenu.cursor, err
}

func (s *MenuService) SelectIssueFormat() error {
	return s.formatterMenu.Select()
}

func (s *MenuService) RegisterMainMenu(mainMenuActions []string) {
	s.mainMenu = &StaticMenu{
		prompt:  "Select an option",
		entries: mainMenuActions,
	}
}

func (s *MenuService) RegisterAdvancedMenu(advancedMenuActions []string) {
	s.advancedMenu = &StaticMenu{
		prompt:  "Select an option",
		entries: advancedMenuActions,
	}
}

func (s *MenuService) RegisterFormatterMenu(formatterMenu *FormatterMenu) {
	s.formatterMenu = formatterMenu
}

func (s *MenuService) RegisterIssueLinkTypeMenu(app *App) {
	s.issueLinkTypeMenu = &IssueLinkTypeMenu{
		jiraClientFactory: app.jiraClientFactory,
		issueLinkTypes:    nil,
		cursor:            0,
	}
}

func (s *MenuService) RegisterIssueSearchMenu(app *App) {
	s.issueSearchMenu = &IssueSearchMenu{
		workbench:           app.workbench,
		issueSelector:       app.issueSelector,
		issueSearchService:  app.issueSearchService,
		workbenchElseGlobal: false,
		cursor:              0,
	}
}

func (s *MenuService) RegisterUserFavoritesMenu(app *App) {
	s.userFavoritesMenu = &UsersFavoritesMenu{
		favorites: app.favoritesService,
	}
}

func NewMenuService(
	config *Config,
) *MenuService {
	return &MenuService{
		config: config,
		jqlMenu: &FzfMenu{
			prompt:  "Please select a JQL",
			entries: keysFromMap(config.JQLs),
		},
	}
}
