// package tinyini provides an extremely bare-bones library for parsing
// INI-like configuration files. For details, see the documentation for
// function Parse.
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
//
// Parse will parse as much as possible even when encountering errors, so
// result may contain something useful even if len(errs) > 0.
//
// The global section is given with the empty section name`""`. Otherwise the
// ection names will be whatever valid UTF-8 is found between the brackets
// `[` and `]`.
//
// Parse ignores whitespace around section headers, keys, and non-quoted
// values. If the value should contain whitespace in its beginning or end,
// enclose the whole value in quotes (`"  value with whitespaces  "`).
//
// Quotes may be contained in quoted values by escaping them like `\"`.
// No quoted expression is handled by Parse, that is, it will return the
// raw value verbatim.
func Parse(r io.Reader) (result map[string]Section, errs []error) {
	s := bufio.NewScanner(r)

	result = map[string]Section{}
	cursection := ""

	akv := func(key, val string) {
		if _, ok := result[cursection]; !ok {
			result[cursection] = Section{}
		}
		result[cursection][key] = append(result[cursection][key], val)
	}

	lineno := 1
	for s.Scan() {
		line := s.Text()
		if m := matcherkeyvalq.FindStringSubmatch(line); m != nil {
			akv(m[1], m[2])
		} else if m := matcherkeyval.FindStringSubmatch(line); m != nil {
			akv(m[1], m[2])
		} else if m := matchersection.FindStringSubmatch(line); m != nil {
			cursection = m[1]
		} else if m := matcherempty.FindStringIndex(line); m != nil {
			// fallthrough
		} else {
			errs = append(errs, newError(lineno, "not section nor key-value"))
		}
		lineno++

	}
	if err := s.Err(); err != nil {
		errs = append(errs, err)
	}
	return
}
