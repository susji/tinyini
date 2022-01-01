// package tinyini provides an extremely bare-bones library for parsing
// INI-like configuration files.
package tinyini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
)

// Section contains all configuration key-values of a single bracketed
// section ("[example-section]"). All key-values may contain multiple
// values. The values are given in the order of occurrence.
type Section map[string][]string

// IniError describes a parsing error and provides its line number.
type IniError struct {
	wrapped error
	Line    int
}

var matchersection = regexp.MustCompile(`^\s*\[(.+?)\]`)
var matcherkeyval = regexp.MustCompile(`^\s*(.+?)\s*=\s*(.*?)\s*$`)
var matcherkeyvalq = regexp.MustCompile(`^\s*(.+?)\s*=\s*"((\\.|[^"\\])*)"`)
var matcherempty = regexp.MustCompile(`^\s*$`)

func (i *IniError) Error() string {
	return fmt.Sprintf("%d: %v", i.Line, i.wrapped)
}

func (i *IniError) Unwrap() error {
	return i.wrapped
}

func newError(line int, msg string) *IniError {
	return &IniError{
		wrapped: errors.New(msg),
		Line:    line,
	}
}

// Parse will produce a map of Section from an io.Reader. The caller should
// note that Parse returns a slice of errors in the order of occurrence, so
// the condition for success is len(errs) == 0.
func Parse(r io.Reader) (result map[string]Section, errs []error) {
	s := bufio.NewScanner(r)
	lines := []string{}
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, []error{err}
	}

	result = map[string]Section{}
	cursection := ""

	akv := func(key, val string) {
		if _, ok := result[cursection]; !ok {
			result[cursection] = Section{}
		}
		result[cursection][key] = append(result[cursection][key], val)
	}

	for i, line := range lines {
		if m := matcherkeyvalq.FindStringSubmatch(line); m != nil {
			akv(m[1], m[2])
		} else if m := matcherkeyval.FindStringSubmatch(line); m != nil {
			akv(m[1], m[2])
		} else if m := matchersection.FindStringSubmatch(line); m != nil {
			cursection = m[1]
		} else if m := matcherempty.FindStringIndex(line); m != nil {
			continue
		} else {
			errs = append(errs, newError(i+1, "not section nor key-value"))
		}
	}
	return
}
