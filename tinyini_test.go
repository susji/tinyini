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
