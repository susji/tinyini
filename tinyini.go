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
	"strings"
)

// Section contains all configuration key-values of a single bracketed
// section ("[example-section]"). All key-values may contain multiple
// values. The values are given in the order of occurrence.
type Section map[string][]Pair

// Sections is a convenience type over map[string]Sections.
type Sections map[string]Section

// ForEacher is the callback function type for iterating over values.
type ForEacher func(section, name, value string) bool

// IniError describes a parsing error and provides its line number.
type IniError struct {
	wrapped error
	Lineno  int
}

type Pair struct {
	Value  string
	Lineno int
}

var matchercomment = regexp.MustCompile(`^\s*;`)
var matchersection = regexp.MustCompile(`^\s*\[(.+?)\]`)
var matcherkeyval = regexp.MustCompile(`^\s*(.+?)\s*=\s*(.*?)\s*(;.*)?$`)
var matcherkeyvalq = regexp.MustCompile(`^\s*(.+?)\s*=\s*"((\\.|[^"\\])*)"\s*(;.*)?$`)
var matcherempty = regexp.MustCompile(`^\s*$`)

func (i *IniError) Error() string {
	return fmt.Sprintf("%d: %v", i.Lineno, i.wrapped)
}

func (i *IniError) Unwrap() error {
	return i.wrapped
}

func newError(lineno int, msg string) *IniError {
	return &IniError{
		wrapped: errors.New(msg),
		Lineno:  lineno,
	}
}

// Parse will produce a map of Section from an io.Reader. The caller should
// note that Parse returns a slice of errors in the order of occurrence, so
// the condition for success is len(errs) == 0.
//
// Parse will parse as much as possible even when encountering errors, so
// result may contain something useful even if len(errs) > 0.
//
// The global section is given with the empty section name "". Otherwise the
// section names will be whatever valid UTF-8 is found between the brackets
// '[' and ']'.
//
// Parse ignores whitespace around section headers, keys, and non-quoted
// values. If the value should contain whitespace in its beginning or end,
// enclose the whole value in quotes ("  value with whitespaces  ").
//
// Quotes may be contained in quoted values by escaping them with the
// backslash like \". Escaped quotes will be unquoted when parsing, but
// all other seemingly "escaped" values like \n are ignored and left verbatim.
//
// All keys may contain multiple values. Their additional values are
// appended to their respective section in the order of appearance.
func Parse(r io.Reader) (result Sections, errs []error) {
	s := bufio.NewScanner(r)

	result = map[string]Section{}
	cursection := ""

	akv := func(lineno int, key, val string) {
		if _, ok := result[cursection]; !ok {
			result[cursection] = Section{}
		}
		result[cursection][key] = append(
			result[cursection][key], Pair{Value: val, Lineno: lineno})
	}

	lineno := 1
	for s.Scan() {
		line := s.Text()
		if m := matchercomment.FindStringIndex(line); m != nil {
			// fallthrough
		} else if m := matcherkeyvalq.FindStringSubmatch(line); m != nil {
			akv(lineno, m[1], strings.Replace(m[2], `\"`, `"`, -1))
		} else if m := matcherkeyval.FindStringSubmatch(line); m != nil {
			akv(lineno, m[1], m[2])
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

// ForEach is a convenience function for simple iteration over Sections. The
// passed callback is called with parsed values in the order of appearance. If
// the callback returns false, iteration is stopped. If a section-specific
// variable has been defined multiple times, each value will be passed on to the
// callback with separate invocations.
func (s Sections) ForEach(callback ForEacher) {
	for sn, section := range s {
		for vn, pairs := range section {
			for _, pair := range pairs {
				if !callback(sn, vn, pair.Value) {
					return
				}
			}
		}
	}
}
