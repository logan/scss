package css3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func readJson(path string, t *testing.T) interface{} {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	b := make([]byte, fi.Size())
	n, err := f.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	b = b[:n]

	var data interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func testParser(s string) *Parser { return NewParser(bytes.NewReader([]byte(s))) }

func simplify(nodes []Node) []interface{} {
	result := make([]interface{}, len(nodes))
	for i, node := range nodes {
		switch node.(type) {
		case EOFNode:
			return result[:i]
		default:
			result[i] = node.TestRepr()
		}
	}
	return result
}

func testJsonSingular(t *testing.T, jsonPath string, parser func(string) Node) {
	_testJson(t, jsonPath, parser, true)
}

func testJson(t *testing.T, jsonPath string, parser func(string) []Node) {
	_testJson(t, jsonPath, parser, false)
}

func _testJson(t *testing.T, jsonPath string, parser interface{}, singular bool) {
	shouldParseInto := func(actual interface{}, expected ...interface{}) string {
		var produced interface{}
		testCase := actual.(string)
		if singular {
			produced = parser.(func(string) Node)(testCase).TestRepr()
		} else {
			produced = simplify(parser.(func(string) []Node)(testCase))
		}
		if msg := ShouldResemble(produced, expected); msg != "" {
			/*
				for i, x := range expected {
					if !reflect.DeepEqual(x, produced.([]interface{})[i]) {
						return fmt.Sprintf("[%d]:\n%#v\n%#v", i, x, produced.([]interface{})[i])
					}
				}
			*/
			pmsg := make(map[string]string)
			if err := json.Unmarshal([]byte(msg), &pmsg); err != nil {
				t.Fatal(err)
			}
			pmsg["Message"] = fmt.Sprintf("%#v\n%s", testCase, pmsg["Message"])
			bs, err := json.Marshal(pmsg)
			if err != nil {
				t.Fatal(err)
			}
			return string(bs)
		}
		return ""
	}

	Convey(jsonPath, t, func() {
		data := readJson(jsonPath, t)

		testSuite := data.([]interface{})
		for i := 0; i < len(testSuite); i += 2 {
			So(testSuite[i], shouldParseInto, testSuite[i+1].([]interface{})...)
		}
	})
}

func TestComponentValueList(t *testing.T) {
	testJson(t, "css-parsing-tests/component_value_list.json",
		func(s string) []Node { return testParser(s).ParseListOfComponentValues() })
}

func TestDeclaration(t *testing.T) {
	testJsonSingular(t, "css-parsing-tests/one_declaration.json",
		func(s string) Node { return testParser(s).ParseDeclaration() })
}

func TestDeclarationList(t *testing.T) {
	testJson(t, "css-parsing-tests/declaration_list.json",
		func(s string) []Node { return testParser(s).ParseDeclarationList() })
}

func TestRule(t *testing.T) {
	testJsonSingular(t, "css-parsing-tests/one_rule.json",
		func(s string) Node { return testParser(s).ParseRule() })
}

func TestRuleList(t *testing.T) {
	testJson(t, "css-parsing-tests/rule_list.json",
		func(s string) []Node { return testParser(s).ParseRuleList() })
}

func TestStylesheet(t *testing.T) {
	testJson(t, "css-parsing-tests/stylesheet.json",
		func(s string) []Node { return testParser(s).ParseStylesheet() })
}
