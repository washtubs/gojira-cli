package cli

import "log"

// Favorites represent entities that the user will prefer to
// interact with frequently,
// and the assumption is that adding such entities
// to a favorites list and searching that list is a preferable UX to a global search
// Favorites are saved in the config file and cached in the favorites service
type FavoritesService struct {
	config            *FavoritesConfig
	jiraClientFactory *JiraClientFactory
}

func (f *FavoritesService) Users() []string {
	if f.config == nil {
		log.Printf("No favorites")
		return []string{}
	}
	return f.config.Users
}

func NewFavoritesService(
	config *FavoritesConfig,
	jiraClientFactory *JiraClientFactory,
) *FavoritesService {
	return &FavoritesService{
		config,
		jiraClientFactory,
	}
}

type UsersFavoritesMenu struct {
	favorites    *FavoritesService
	cursor       int
	selectedUser string
}

func (m *UsersFavoritesMenu) SelectedUser() string {
	return m.selectedUser
}

func (m *UsersFavoritesMenu) Select(prompt string) error {
	users := m.favorites.Users()
	formatters := make([]Formatter, len(users))
	for i, u := range users {
		formatters[i] = StringFormatter(u)
	}
	idxs, cancelled, err := FzfSelect(formatters, SelectOptions{Prompt: prompt, One: true}, 0)
	if cancelled {
		return CancelError()
	}
	if err != nil {
		return err
	}
	if len(idxs) != 1 {
		panic("expected exactly one")
	}
	m.cursor = idxs[0]
	m.selectedUser = users[m.cursor]
	return nil
}
