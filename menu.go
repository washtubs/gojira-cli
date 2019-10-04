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
		log.Printf("Error making selection: ", err)
	}
	return err
}

// Basically mediates access to the simple promts using promptui
type MenuService struct {
	config            *Config
	jqlMenu           *StaticMenu
	mainMenu          *StaticMenu
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
		jqlMenu: &StaticMenu{
			prompt:  "Please select a JQL",
			entries: keysFromMap(config.JQLs),
		},
	}
}
