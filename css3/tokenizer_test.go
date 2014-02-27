package css3

import (
	"bytes"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToken(t *testing.T) {
	Convey("Token to string", t, func() {
		for i := MinTokenType; i <= MaxTokenType+1; i++ {
			So(Token{i, "test"}.String(), ShouldEqual, "{"+i.String()+"=test}")
		}
	})
}

func TestNumeric(t *testing.T) {
	Convey("Numeric to string", t, func() {
		So(Numeric{Repr: "1"}.String(), ShouldEqual, "<num(int):1>")
		So(Numeric{NumberType: Float, Repr: "1.0"}.String(),
			ShouldEqual, "<num(float):1.0>")
		So(Numeric{Repr: "1", Unit: "px"}.String(), ShouldEqual, "<num(int):1px>")
		So(Numeric{NumberType: Float, Repr: "1.0", Unit: "%"}.String(),
			ShouldEqual, "<num(float):1.0%>")
	})
}

func TestConsumeToken(t *testing.T) {
	newTokenizer := func(input string) *Tokenizer {
		return NewTokenizer(bytes.NewReader([]byte(input)))
	}

	shouldYield := func(actual interface{}, expected ...interface{}) string {
		tokenizer := newTokenizer(actual.(string))
		for _, exp := range expected {
			tok, err := tokenizer.ConsumeToken()
			if msg := ShouldBeNil(err); msg != "" {
				return msg
			}
			if msg := ShouldResemble(tok, exp); msg != "" {
				return msg
			}
		}
		return ""
	}

	Convey(`U+0022 QUOTATION MARK (")`, t, func() {
		So(`"`, shouldYield, NewToken(StringToken, ""))
		So(`"test`, shouldYield, NewToken(StringToken, "test"))
		So(`"test"test`, shouldYield, NewToken(StringToken, "test"))
		So(`"\"test\""`, shouldYield, NewToken(StringToken, `"test"`))
		So("\"test\n\"", shouldYield, NewToken(BadStringToken, "test"))
		So(`"\2318"`, shouldYield, NewToken(StringToken, "\u2318"))
		So(`"\002318ff"`, shouldYield, NewToken(StringToken, "\u2318ff"))
		So(`"\\0022 is \0022"`, shouldYield, NewToken(StringToken, `\0022 is "`))
		So(`"\2318`, shouldYield, NewToken(StringToken, "\u2318"))
		So(`"\0 test`, shouldYield, NewToken(StringToken, "\ufffdtest"))
		So(`"test\`, shouldYield, NewToken(StringToken, "test"))
		So(`"\2A"`, shouldYield, NewToken(StringToken, "*"))
		So("\"\\\ntest\\\n\"", shouldYield, NewToken(StringToken, "test"))
	})

	Convey("U+0023 NUMBER SIGN (#)", t, func() {
		So("#", shouldYield, NewDelimToken('#'))
		So("#abc", shouldYield, NewToken(HashToken, "abc"))
		So("#123abc", shouldYield, NewToken(HashToken, "123abc"))
		So("#\\\n", shouldYield, NewDelimToken('#'))
		So(`#\`, shouldYield, NewToken(HashToken, "\ufffd"))
		So("#=", shouldYield, NewDelimToken('#'))
	})

	Convey("U+0024 DOLLAR SIGN ($)", t, func() {
		So("$", shouldYield, NewDelimToken('$'))
		So("$.", shouldYield, NewDelimToken('$'))
		So("$=", shouldYield, NewToken(SuffixMatchToken, nil))
	})

	Convey("U+0027 APOSTROPHE (')", t, func() {
		So("'test'", shouldYield, NewToken(StringToken, "test"))
	})

	Convey("U+0028 LEFT PARENTHESIS (()", t, func() {
		So("(", shouldYield, NewToken(LParenToken, nil))
	})

	Convey("U+0029 RIGHT PARENTHESIS ())", t, func() {
		So(")", shouldYield, NewToken(RParenToken, nil))
	})

	Convey("U+002A ASTERISK (*)", t, func() {
		So("*", shouldYield, NewDelimToken('*'))
		So("*.", shouldYield, NewDelimToken('*'))
		So("*=", shouldYield, NewToken(SubstringMatchToken, nil))
	})

	Convey("U+002B PLUS SIGN (+)", t, func() {
		So("+", shouldYield, NewDelimToken('+'))
		So("+1", shouldYield, NewToken(NumberToken, &Numeric{Repr: "+1", Integer: 1}))
		So("+a", shouldYield, NewDelimToken('+'))
	})

	Convey("U+002C COMMA (,)", t, func() {
		So(",", shouldYield, NewToken(CommaToken, nil))
	})

	Convey("U+002D MINUS (-)", t, func() {
		So("-", shouldYield, NewDelimToken('-'))
		So("-1", shouldYield, NewToken(NumberToken, &Numeric{Repr: "-1", Integer: -1}))
		So("-a", shouldYield, NewToken(IdentToken, "-a"))
		So("--->", shouldYield, NewDelimToken('-'))
		So("-->", shouldYield, NewToken(CDCToken, nil))
		So("->", shouldYield, NewDelimToken('-'))
	})

	Convey("U+002E FULL STOP (.)", t, func() {
		So(".", shouldYield, NewDelimToken('.'))
		So("..", shouldYield, NewDelimToken('.'))
		So(".0", shouldYield, NewToken(NumberToken,
			&Numeric{NumberType: Float, Repr: ".0", Float: 0}))
		So(".1e2", shouldYield, NewToken(NumberToken,
			&Numeric{NumberType: Float, Repr: ".1e2", Float: 10}))
	})

	Convey("U+002F SOLIDUS (/)", t, func() {
		So("/", shouldYield, NewDelimToken('/'))
		So("//", shouldYield, NewDelimToken('/'))
		So("/*/123", shouldYield, NewEOFToken())
		So("/**/123", shouldYield,
			NewToken(NumberToken, &Numeric{Repr: "123", Integer: 123}))
		So("/** test **//** test **/123", shouldYield,
			NewToken(NumberToken, &Numeric{Repr: "123", Integer: 123}))
	})

	Convey("U+003A COLON (:)", t, func() {
		So(":", shouldYield, NewToken(ColonToken, nil))
	})

	Convey("U+003B SEMICOLON (;)", t, func() {
		So(";", shouldYield, NewToken(SemicolonToken, nil))
	})

	Convey("U+003C LESS-THAN SIGN (<)", t, func() {
		So("<", shouldYield, NewDelimToken('<'))
		So("<!", shouldYield, NewDelimToken('<'))
		So("<!-", shouldYield, NewDelimToken('<'))
		So("<!--", shouldYield, NewToken(CDOToken, nil))
	})

	Convey("U+0040 COMMERCIAL AT (@)", t, func() {
		So("@", shouldYield, NewDelimToken('@'))
		So("@-", shouldYield, NewDelimToken('@'))
		So("@-\\\n", shouldYield, NewDelimToken('@'))
		So("@-t", shouldYield, NewToken(AtKeywordToken, "-t"))
		So(`@-\`, shouldYield, NewToken(AtKeywordToken, "-\ufffd"))
		So("@test", shouldYield, NewToken(AtKeywordToken, "test"))
	})

	Convey("U+005B LEFT SQUARE BRACKET ([)", t, func() {
		So("[", shouldYield, NewToken(LSquareToken, nil))
	})

	Convey("U+005D RIGHT SQUARE BRACKET (])", t, func() {
		So("]", shouldYield, NewToken(RSquareToken, nil))
	})

	Convey("U+005E CIRCUMFLEX ACCENT (^)", t, func() {
		So("^", shouldYield, NewDelimToken('^'))
		So("^.", shouldYield, NewDelimToken('^'))
		So("^=", shouldYield, NewToken(PrefixMatchToken, nil))
	})

	Convey("U+007B LEFT CURLY BRACKET ({)", t, func() {
		So("{", shouldYield, NewToken(LCurlyToken, nil))
	})

	Convey("U+007C VERTICAL LINE (|)", t, func() {
		So("|", shouldYield, NewDelimToken('|'))
		So("|.", shouldYield, NewDelimToken('|'))
		So("|=", shouldYield, NewToken(DashMatchToken, nil))
		So("||", shouldYield, NewToken(ColumnToken, nil))
	})

	Convey("U+007D RIGHT CURLY BRACKET (})", t, func() {
		So("}", shouldYield, NewToken(RCurlyToken, nil))
	})

	Convey("U+007E TILDE (~)", t, func() {
		So("~", shouldYield, NewDelimToken('~'))
		So("~.", shouldYield, NewDelimToken('~'))
		So("~=", shouldYield, NewToken(IncludeMatchToken, nil))
	})

	Convey("number", t, func() {
		intNum := func(repr string) *Numeric {
			num := &Numeric{NumberType: Integer, Repr: repr}
			num.Integer, _ = strconv.ParseInt(repr, 0, 64)
			return num
		}

		floatNum := func(repr string) *Numeric {
			num := &Numeric{NumberType: Float, Repr: repr}
			num.Float, _ = strconv.ParseFloat(repr, 64)
			return num
		}

		dim := func(num *Numeric, unit string) *Numeric {
			num.Unit = unit
			return num
		}

		So("1", shouldYield, NewToken(NumberToken, intNum("1")))
		So("1.0", shouldYield, NewToken(NumberToken, floatNum("1.0")))

		So("+1234567890em", shouldYield,
			NewToken(DimensionToken, dim(intNum("+1234567890"), "em")))
		So("-12345.67890px", shouldYield,
			NewToken(DimensionToken, dim(floatNum("-12345.67890"), "px")))

		So("1e1", shouldYield, NewToken(NumberToken, floatNum("1e1")))
		So("1e2%", shouldYield, NewToken(PercentageToken, dim(floatNum("1e2"), "%")))
		So("1.2e3.4", shouldYield, NewToken(NumberToken, floatNum("1.2e3")))
		So("-1e-1", shouldYield, NewToken(NumberToken, floatNum("-1e-1")))
		So("-1e+2em", shouldYield, NewToken(DimensionToken, dim(floatNum("-1e+2"), "em")))
		So(`1e\m`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "em")))
		So(`1\65\6d`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "em")))
		So(`1\000025`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "%")))
		So(`1\d888`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "\ufffd")))
		So(`1\110000`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "\ufffd")))
		So(`1\`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "\ufffd")))
		So("1e\\\n", shouldYield, NewToken(DimensionToken, dim(intNum("1"), "e")))
		So(`1test\`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "test\ufffd")))
		So("1-x", shouldYield, NewToken(DimensionToken, dim(intNum("1"), "-x")))
		So(`1-\`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "-\ufffd")))
		So(`1-\n`, shouldYield, NewToken(DimensionToken, dim(intNum("1"), "-n")))
		So("1-\\\n", shouldYield, NewToken(NumberToken, intNum("1")))

		tk := newTokenizer("1e900")
		_, err := tk.ConsumeToken()
		So(err, ShouldNotBeNil)
	})

	Convey("identifier", t, func() {
		So("test", shouldYield, NewToken(IdentToken, "test"))
		So("test ing", shouldYield, NewToken(IdentToken, "test"))
		So("test(ing)", shouldYield, NewToken(FunctionToken, "test"))
		So(`\'test\'`, shouldYield, NewToken(IdentToken, "'test'"))
		So("\\\n", shouldYield, NewDelimToken('\\'))

		So("url(ing)", shouldYield, NewToken(UrlToken, "ing"))
		So("url(", shouldYield, NewToken(UrlToken, ""))
		So("url( ", shouldYield, NewToken(UrlToken, ""))
		So("url()", shouldYield, NewToken(UrlToken, ""))
		So("url(test)", shouldYield, NewToken(UrlToken, "test"))
		So("url(  test  )", shouldYield, NewToken(UrlToken, "test"))
		So("url('test')", shouldYield, NewToken(UrlToken, "test"))
		So("url(  'test'  )", shouldYield, NewToken(UrlToken, "test"))
		So(`url('test\'test')`, shouldYield, NewToken(UrlToken, "test'test"))
		So(`url(\'test\')`, shouldYield, NewToken(UrlToken, "'test'"))
		So("url('", shouldYield, NewToken(UrlToken, ""))
		So("url(''   ,", shouldYield, NewToken(BadUrlToken, nil))
		So("url((", shouldYield, NewToken(BadUrlToken, nil))
		So("url(x'", shouldYield, NewToken(BadUrlToken, nil))
		So("url(x\\\n", shouldYield, NewToken(BadUrlToken, nil))
		So("url(x \nx", shouldYield, NewToken(BadUrlToken, nil))
		So("url('x\n')", shouldYield, NewToken(BadUrlToken, nil))
		So("url('x\n\\)x)y", shouldYield, NewToken(BadUrlToken, nil), NewToken(IdentToken, "y"))
		So("url(\001)", shouldYield, NewToken(BadUrlToken, nil))
		So("url(\\001)", shouldYield, NewToken(UrlToken, "\001"))
	})

	Convey("unicode range", t, func() {
		urange := func(start, end rune) *Token {
			return NewToken(UnicodeRangeToken, UnicodeRange{start, end})
		}

		So("u+u", shouldYield, NewToken(IdentToken, "u"))
		So("u+0", shouldYield, urange(0, 0))
		So("u+00100?", shouldYield, urange(0x1000, 0x100f))
		So("u+001???", shouldYield, urange(0x1000, 0x1fff))
		So("u+001000?", shouldYield, urange(0x1000, 0x1000))
		So("u+1000-1011", shouldYield, urange(0x1000, 0x1011))
		So("u+1000-101?", shouldYield, urange(0x1000, 0x101))
		So("u+100?-1011", shouldYield, urange(0x1000, 0x100f))
	})

	Convey("EOF", t, func() {
		So("", shouldYield, NewEOFToken())
	})

	Convey("anything else", t, func() {
		So("!", shouldYield, NewDelimToken('!'))
	})
}
