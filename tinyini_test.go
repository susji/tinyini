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
	if len(res) != 3 {
		t.Error("expecting 3 sections, got ", len(res))
	}

	if res[""] == nil {
		t.Error("missing global section")
	}
	if res["section"] == nil {
		t.Error("missing section")
	}
	if !reflect.DeepEqual(res[""]["globalkey"], []string{"globalvalue"}) {
		t.Errorf("unexpected global value: %#v", res[""]["globalkey"])
	}
	if !reflect.DeepEqual(res["section"]["key"], []string{"first-value", "second-value"}) {
		t.Error("missing sectioned values")
	}
	if !reflect.DeepEqual(res["section"]["empty"], []string{""}) {
		t.Error("missing empty value")
	}
	if !reflect.DeepEqual(res["section"]["anotherkey"], []string{"  has whitespace   "}) {
		t.Error("missing quoted value")
	}
	if res["änöther-section"] == nil {
		t.Error("missing änöther-section")
	}
	if !reflect.DeepEqual(res["änöther-section"]["key"], []string{"different value"}) {
		t.Error("unexpected änöther-sectioned value")
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
