package tinyini_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	ti "github.com/susji/tinyini"
)

func TestBasic(t *testing.T) {
	c := `
globalkey = globalvalue

[section]
key = first-value
key = second-value
empty= ;ends with a comment
; comment line
  ; another comment line
anotherkey = "  has whitespace   " ; ends with a comment

[änöther-section] ; this is a comment and ignored
key = "different value"
`
	res, errs := ti.Parse(strings.NewReader(c))
	if len(errs) > 0 {
		t.Fatalf("should have no error, got %v", errs)
	}
	if !reflect.DeepEqual(
		res,
		map[string]ti.Section{
			"": ti.Section{"globalkey": []ti.Pair{ti.Pair{"globalvalue", 2}}},
			"section": ti.Section{
				"key":        []ti.Pair{ti.Pair{"first-value", 5}, ti.Pair{"second-value", 6}},
				"empty":      []ti.Pair{ti.Pair{"", 7}},
				"anotherkey": []ti.Pair{ti.Pair{"  has whitespace   ", 10}},
			},
			"änöther-section": ti.Section{
				"key": []ti.Pair{ti.Pair{"different value", 13}}},
		}) {
		t.Errorf("missing sectioned values, got %#v", res["section"])
	}
}

func TestError(t *testing.T) {
	table := []struct {
		conf   string
		lineno int
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
		t.Run(fmt.Sprintf("%s_%d", entry.conf, entry.lineno), func(t *testing.T) {
			_, errs := ti.Parse(strings.NewReader(entry.conf))
			if len(errs) != 1 {
				t.Errorf("expecting 1 error, got %d", len(errs))
				return
			}
			err := errs[0].(*ti.IniError)
			if err.Lineno != entry.lineno {
				t.Errorf("error line %d, wanted %d",
					err.Lineno,
					entry.lineno)
			}
		})
	}
}

func TestQuoted(t *testing.T) {
	// We do not attempt to separate "almost quoted" from
	// "properly quoted" values. This test reflects that.
	table := []struct {
		give string
		want string
	}{
		{
			`key = "value"`,
			"value",
		},
		{
			`key = "value" ; comment`,
			"value",
		},
		{
			`key = "\a\b\n"`,
			`\a\b\n`,
		},
		{
			`key = "`,
			`"`,
		},
		{
			`key = "hola\"`,
			`"hola\"`,
		},
		{
			`key = "\"value\""`,
			`"value"`,
		},
		{
			`key = "\\\"value\\\""`,
			`\\"value\\"`,
		},
	}
	for _, entry := range table {
		t.Run(fmt.Sprintf("%s_%v", entry.give, entry.want), func(t *testing.T) {
			got, errs := ti.Parse(strings.NewReader(entry.give))
			if len(errs) != 0 {
				t.Errorf("expecting no errors, got %d", len(errs))
			}
			if !reflect.DeepEqual(
				got,
				map[string]ti.Section{
					"": ti.Section{"key": []ti.Pair{ti.Pair{entry.want, 1}}}}) {
				t.Errorf("got %#v, want %#v", got, entry.want)
			}
		})
	}
}
