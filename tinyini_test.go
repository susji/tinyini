package tinyini_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/susji/tinyini"
)

func TestBasic(t *testing.T) {
	c := `
globalkey = globalvalue

[section]
key = first-value
key = second-value
empty=
anotherkey = "  has whitespace   "

[änöther-section] ; this is a comment and ignored
key = different value
`
	res, errs := tinyini.Parse(strings.NewReader(c))
	if len(errs) > 0 {
		t.Fatalf("should have no error, got %v", errs)
	}
	if !reflect.DeepEqual(
		res,
		map[string]tinyini.Section{
			"": tinyini.Section{"globalkey": []string{"globalvalue"}},
			"section": tinyini.Section{
				"key":        []string{"first-value", "second-value"},
				"empty":      []string{""},
				"anotherkey": []string{"  has whitespace   "},
			},
			"änöther-section": tinyini.Section{
				"key": []string{"different value"}},
		}) {
		t.Errorf("missing sectioned values, got %#v", res["section"])
	}
}

func TestError(t *testing.T) {
	table := []struct {
		conf string
		line int
	}{
		{`ok = value
error
`, 2},
		{`[section]
[another-section]
[borken
`, 3},
	}

	for _, entry := range table {
		t.Run(fmt.Sprintf("%s_%d", entry.conf, entry.line), func(t *testing.T) {
			_, errs := tinyini.Parse(strings.NewReader(entry.conf))
			if len(errs) != 1 {
				t.Errorf("expecting 1 error, got %d", len(errs))
				return
			}
			err := errs[0].(*tinyini.IniError)
			if err.Line != entry.line {
				t.Errorf("error line %d, wanted %d",
					err.Line,
					entry.line)
			}
		})
	}
}

func TestQuoted(t *testing.T) {
	// We do not attempt to separate "almost quoted" from
	// "properly quoted" values. This test reflects that.
	table := []struct {
		give string
		want tinyini.Section
	}{
		{
			`key = "value"`,
			tinyini.Section{"key": []string{"value"}},
		},
		{
			`key = "\a\b\n"`,
			tinyini.Section{"key": []string{`\a\b\n`}},
		},
		{
			`key = "`,
			tinyini.Section{"key": []string{`"`}},
		},
		{
			`key = "hola\"`,
			tinyini.Section{"key": []string{`"hola\"`}},
		},
		{
			`key = "\"value\""`,
			tinyini.Section{"key": []string{`"value"`}},
		},
		{
			`key = "\\\"value\\\""`,
			tinyini.Section{"key": []string{`\\"value\\"`}},
		},
	}
	for _, entry := range table {
		t.Run(fmt.Sprintf("%s_%v", entry.give, entry.want), func(t *testing.T) {
			got, errs := tinyini.Parse(strings.NewReader(entry.give))
			if len(errs) != 0 {
				t.Errorf("expecting no errors, got %d", len(errs))
			}
			if !reflect.DeepEqual(got, map[string]tinyini.Section{"": entry.want}) {
				t.Errorf("got %#v, want %#v", got, entry.want)
			}
		})
	}
}
