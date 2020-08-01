package cli

type Formatter interface {
	Format() string
}

type StringFormatter string

func (s StringFormatter) Format() string {
	return string(s)
}
