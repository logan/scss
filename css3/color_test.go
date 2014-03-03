package css3

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestColors(t *testing.T) {
	shouldParseInto := func(actual interface{}, expected ...interface{}) (msg string) {
		testCase := actual.(string)
		color := ColorFromString(testCase)
		if expected[0] == nil {
			msg = ShouldBeNil(color)
		} else if str, ok := expected[0].(string); ok {
			msg = ShouldResemble(color.TestRepr(), str)
		} else if nums, ok := expected[0].([]interface{}); ok {
			codes := make([]float64, len(nums))
			for i, num := range nums {
				switch n := num.(type) {
				case int:
					codes[i] = float64(n)
				case float64:
					codes[i] = n
				default:
					t.Fatal("don't know how to parse expected value of %#v", nums)
				}
			}
			msg = ShouldResemble(color.TestRepr(), codes)
			if msg != "" {
				ok := ShouldAlmostEqual(color.R, codes[0]) == ""
				ok = ok && ShouldAlmostEqual(color.G, codes[1]) == ""
				ok = ok && ShouldAlmostEqual(color.B, codes[2]) == ""
				ok = ok && ShouldAlmostEqual(color.A, codes[3]) == ""
				if ok {
					return ""
				}
			}
		} else {
			t.Fatal("don't know how to parse expected value of %#v", expected[0])
		}
		if msg != "" {
			pmsg := make(map[string]string)
			if err := json.Unmarshal([]byte(msg), &pmsg); err != nil {
				return fmt.Sprintf("%#v\n%s", testCase, msg)
			}
			pmsg["Message"] = fmt.Sprintf("%#v\n%s", testCase, pmsg["Message"])
			bs, err := json.Marshal(pmsg)
			if err != nil {
				t.Fatal(err)
			}
			msg = string(bs)
		}
		return
	}

	Convey("Colors", t, func() {
		data := readJson("css-parsing-tests/color3.json", t)
		testSuite := data.([]interface{})
		for i := 0; i < len(testSuite); i += 2 {
			So(testSuite[i], shouldParseInto, testSuite[i+1])
		}
	})
}
