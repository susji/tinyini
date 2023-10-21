package tinyini_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/susji/tinyini"
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
		ti.Sections{
			"": ti.Section{"globalkey": []ti.Pair{ti.Pair{"globalvalue", 2}}},
			"section": ti.Section{
				"key":        []ti.Pair{ti.Pair{"first-value", 5}, ti.Pair{"second-value", 6}},
				"empty":      []ti.Pair{ti.Pair{"", 7}},
				"anotherkey": []ti.Pair{ti.Pair{"  has whitespace   ", 10}},
			},
			"änöther-section": ti.Section{
				"key": []ti.Pair{ti.Pair{"different value", 13}}},
		}) {
		t.Errorf("missing sectioned values, got %#v", res)
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
		{`[section]
onlykey
`, 2},
		{`[section]
key = value
key ;broken
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
				ti.Sections{
					"": ti.Section{"key": []ti.Pair{ti.Pair{entry.want, 1}}}}) {
				t.Errorf("got %#v, want %#v", got, entry.want)
			}
		})
	}
}

func TestForEach(t *testing.T) {
	config := `
toplevelvar1 = tl1
toplevelvar2 = " tl2 "

[section1]
section1var=section1val1
section1var=section1val2

[section2]
section2var=section2val
section2var=notseen
`
	sections, errs := ti.Parse(strings.NewReader(config))
	if len(errs) != 0 {
		t.Errorf("expecting no errors, got %d", len(errs))
	}

	s1var := 0
	sections.ForEach(func(section, key, value string) bool {
		switch section {
		case "":
			switch key {
			case "toplevelvar1":
				if value != "tl1" {
					t.Error("toplevelvar1")
				}
			case "toplevelvar2":
				if value != " tl2 " {
					t.Error("toplevelvar2")
				}
			default:
				t.Error("unrecognized top level definition")
			}
		case "section1":
			switch key {
			case "section1var":
				if s1var == 0 {
					if value != "section1val1" {
						t.Error("section1var#0")
					}
					s1var++
				} else if s1var == 1 {
					if value != "section1val2" {
						t.Error("section1var#1")
					}
					s1var++
				} else {
					t.Error("too many section1val2")
				}

			default:
				t.Error("unrecognized section1 definition")
			}
		case "section2":
			switch key {
			case "section2var":
				if value != "section2val" {
					t.Error("section2var")
				}
				return false
			default:
				t.Error("unrecognized section2 definition")
			}
		}
		return true
	})
}

func ExampleForEach() {
	config := `
value=123
another_value=" xyz "

[section]
value=321
another_value="abc"
`
	sections, errs := tinyini.Parse(strings.NewReader(config))
	if len(errs) > 0 {
		fmt.Println("parsing errors: ", errs)
		os.Exit(1)
	}

	var value int
	var anotherValue string
	sections.ForEach(func(section, k, v string) bool {
		switch section {
		case "":
			switch k {
			case "value":
				value, _ = strconv.Atoi(v)
			}
		case "section":
			switch k {
			case "another_value":
				anotherValue = v
			}
		}
		return true
	})

	fmt.Printf("value is %d and anotherValue is %s\n", value, anotherValue)
	// Output: value is 123 and anotherValue is abc
}
