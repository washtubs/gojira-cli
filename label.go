package cli

type Label string

func (l Label) Format() string {
	return string(l)
}
