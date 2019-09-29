package cli

type Config struct {
	// Jira doesn't seem to keep a list of existing labels so I gotta add them via config
	LabelsAllowed []Label
	// WOuld prefer to use the saved JQL queries for the user, but I
	JQLs map[string]string
}

type ConfigLoader interface {
	LoadConfig() *Config
}

type stubConfigLoader struct{}

func (cl stubConfigLoader) LoadConfig() *Config {
	return &Config{
		[]Label{Label("fooLabel"), Label("barLabel"), Label("bazLabel")},
		map[string]string{
			"foo": "FOO",
			"bar": "BAR",
			"all": "ALL",
		},
	}
}

func NewConfigLoader() ConfigLoader {
	return stubConfigLoader{}
}
