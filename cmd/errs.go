package cmd

import "strings"

type errs []error

func (errs errs) Error() string {
	return errs.Join("\n")
}

func (errs errs) Join(sep string) string {
	switch len(errs) {
	case 0:
		return ""
	case 1:
		return errs[0].Error()
	}

	var b strings.Builder
	b.WriteString(errs[0].Error())
	for _, s := range errs[1:] {
		b.WriteString(sep)
		b.WriteString(s.Error())
	}
	return b.String()
}
