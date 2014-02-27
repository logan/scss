package css3

import (
	"bytes"
	"io"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToken(t *testing.T) {
	Convey("Token to string", t, func() {
		So(Token{0, nil}.String(), ShouldEqual, "{ErrorToken=<nil>}")
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
	shouldTokenize := func(actual interface{}, expected ...interface{}) string {
		scanner, ok := actual.(io.RuneScanner)
		if !ok {
			scanner = bytes.NewReader([]byte(actual.(string)))
		}
		tokenizer := NewTokenizer(scanner)
		for _, exp := range expected {
			tok := tokenizer.ConsumeToken()
			if msg := ShouldResemble(tok, exp); msg != "" {
				return msg
			}
		}
		return ""
	}

	Convey("error handling", t, func() {
		t := func(input string) io.RuneScanner {
			return &testRuneScanner{RuneScanner: bytes.NewReader([]byte(input))}
		}

		So(t("!abc"), shouldTokenize,
			NewErrorToken(testRuneScannerError), NewErrorToken(testRuneScannerError))
	})

	Convey(`U+0022 QUOTATION MARK (")`, t, func() {
		So(`"`, shouldTokenize, NewToken(StringToken, ""))
		So(`"test`, shouldTokenize, NewToken(StringToken, "test"))
		So(`"test"test`, shouldTokenize, NewToken(StringToken, "test"))
		So(`"\"test\""`, shouldTokenize, NewToken(StringToken, `"test"`))
		So("\"test\n\"", shouldTokenize, NewToken(BadStringToken, "test"))
		So(`"\2318"`, shouldTokenize, NewToken(StringToken, "\u2318"))
		So(`"\002318ff"`, shouldTokenize, NewToken(StringToken, "\u2318ff"))
		So(`"\\0022 is \0022"`, shouldTokenize, NewToken(StringToken, `\0022 is "`))
		So(`"\2318`, shouldTokenize, NewToken(StringToken, "\u2318"))
		So(`"\0 test`, shouldTokenize, NewToken(StringToken, "\ufffdtest"))
		So(`"test\`, shouldTokenize, NewToken(StringToken, "test"))
		So(`"\2A"`, shouldTokenize, NewToken(StringToken, "*"))
		So("\"\\\ntest\\\n\"", shouldTokenize, NewToken(StringToken, "test"))
	})

	Convey("U+0023 NUMBER SIGN (#)", t, func() {
		So("#", shouldTokenize, NewDelimToken('#'))
		So("#abc", shouldTokenize, NewToken(HashToken, "abc"))
		So("#123abc", shouldTokenize, NewToken(HashToken, "123abc"))
		So("#\\\n", shouldTokenize, NewDelimToken('#'))
		So(`#\`, shouldTokenize, NewToken(HashToken, "\ufffd"))
		So("#=", shouldTokenize, NewDelimToken('#'))
	})

	Convey("U+0024 DOLLAR SIGN ($)", t, func() {
		So("$", shouldTokenize, NewDelimToken('$'))
		So("$.", shouldTokenize, NewDelimToken('$'))
		So("$=", shouldTokenize, NewToken(SuffixMatchToken, nil))
	})

	Convey("U+0027 APOSTROPHE (')", t, func() {
		So("'test'", shouldTokenize, NewToken(StringToken, "test"))
	})

	Convey("U+0028 LEFT PARENTHESIS (()", t, func() {
		So("(", shouldTokenize, NewToken(LParenToken, nil))
	})

	Convey("U+0029 RIGHT PARENTHESIS ())", t, func() {
		So(")", shouldTokenize, NewToken(RParenToken, nil))
	})

	Convey("U+002A ASTERISK (*)", t, func() {
		So("*", shouldTokenize, NewDelimToken('*'))
		So("*.", shouldTokenize, NewDelimToken('*'))
		So("*=", shouldTokenize, NewToken(SubstringMatchToken, nil))
	})

	Convey("U+002B PLUS SIGN (+)", t, func() {
		So("+", shouldTokenize, NewDelimToken('+'))
		So("+1", shouldTokenize, NewToken(NumberToken, &Numeric{Repr: "+1", Integer: 1}))
		So("+a", shouldTokenize, NewDelimToken('+'))
	})

	Convey("U+002C COMMA (,)", t, func() {
		So(",", shouldTokenize, NewToken(CommaToken, nil))
	})

	Convey("U+002D MINUS (-)", t, func() {
		So("-", shouldTokenize, NewDelimToken('-'))
		So("-1", shouldTokenize, NewToken(NumberToken, &Numeric{Repr: "-1", Integer: -1}))
		So("-a", shouldTokenize, NewToken(IdentToken, "-a"))
		So("--->", shouldTokenize, NewDelimToken('-'))
		So("-->", shouldTokenize, NewToken(CDCToken, nil))
		So("->", shouldTokenize, NewDelimToken('-'))
	})

	Convey("U+002E FULL STOP (.)", t, func() {
		So(".", shouldTokenize, NewDelimToken('.'))
		So("..", shouldTokenize, NewDelimToken('.'))
		So(".0", shouldTokenize, NewToken(NumberToken,
			&Numeric{NumberType: Float, Repr: ".0", Float: 0}))
		So(".1e2", shouldTokenize, NewToken(NumberToken,
			&Numeric{NumberType: Float, Repr: ".1e2", Float: 10}))
	})

	Convey("U+002F SOLIDUS (/)", t, func() {
		So("/", shouldTokenize, NewDelimToken('/'))
		So("//", shouldTokenize, NewDelimToken('/'))
		So("/*/123", shouldTokenize, NewEOFToken())
		So("/**/123", shouldTokenize,
			NewToken(NumberToken, &Numeric{Repr: "123", Integer: 123}))
		So("/** test **//** test **/123", shouldTokenize,
			NewToken(NumberToken, &Numeric{Repr: "123", Integer: 123}))
	})

	Convey("U+003A COLON (:)", t, func() {
		So(":", shouldTokenize, NewToken(ColonToken, nil))
	})

	Convey("U+003B SEMICOLON (;)", t, func() {
		So(";", shouldTokenize, NewToken(SemicolonToken, nil))
	})

	Convey("U+003C LESS-THAN SIGN (<)", t, func() {
		So("<", shouldTokenize, NewDelimToken('<'))
		So("<!", shouldTokenize, NewDelimToken('<'))
		So("<!-", shouldTokenize, NewDelimToken('<'))
		So("<!--", shouldTokenize, NewToken(CDOToken, nil))
	})

	Convey("U+0040 COMMERCIAL AT (@)", t, func() {
		So("@", shouldTokenize, NewDelimToken('@'))
		So("@-", shouldTokenize, NewDelimToken('@'))
		So("@-\\\n", shouldTokenize, NewDelimToken('@'))
		So("@-t", shouldTokenize, NewToken(AtKeywordToken, "-t"))
		So(`@-\`, shouldTokenize, NewToken(AtKeywordToken, "-\ufffd"))
		So("@test", shouldTokenize, NewToken(AtKeywordToken, "test"))
	})

	Convey("U+005B LEFT SQUARE BRACKET ([)", t, func() {
		So("[", shouldTokenize, NewToken(LSquareToken, nil))
	})

	Convey("U+005D RIGHT SQUARE BRACKET (])", t, func() {
		So("]", shouldTokenize, NewToken(RSquareToken, nil))
	})

	Convey("U+005E CIRCUMFLEX ACCENT (^)", t, func() {
		So("^", shouldTokenize, NewDelimToken('^'))
		So("^.", shouldTokenize, NewDelimToken('^'))
		So("^=", shouldTokenize, NewToken(PrefixMatchToken, nil))
	})

	Convey("U+007B LEFT CURLY BRACKET ({)", t, func() {
		So("{", shouldTokenize, NewToken(LCurlyToken, nil))
	})

	Convey("U+007C VERTICAL LINE (|)", t, func() {
		So("|", shouldTokenize, NewDelimToken('|'))
		So("|.", shouldTokenize, NewDelimToken('|'))
		So("|=", shouldTokenize, NewToken(DashMatchToken, nil))
		So("||", shouldTokenize, NewToken(ColumnToken, nil))
	})

	Convey("U+007D RIGHT CURLY BRACKET (})", t, func() {
		So("}", shouldTokenize, NewToken(RCurlyToken, nil))
	})

	Convey("U+007E TILDE (~)", t, func() {
		So("~", shouldTokenize, NewDelimToken('~'))
		So("~.", shouldTokenize, NewDelimToken('~'))
		So("~=", shouldTokenize, NewToken(IncludeMatchToken, nil))
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

		So("1", shouldTokenize, NewToken(NumberToken, intNum("1")))
		So("1.0", shouldTokenize, NewToken(NumberToken, floatNum("1.0")))

		So("+1234567890em", shouldTokenize,
			NewToken(DimensionToken, dim(intNum("+1234567890"), "em")))
		So("-12345.67890px", shouldTokenize,
			NewToken(DimensionToken, dim(floatNum("-12345.67890"), "px")))

		So("1e1", shouldTokenize, NewToken(NumberToken, floatNum("1e1")))
		So("1e2%", shouldTokenize, NewToken(PercentageToken, dim(floatNum("1e2"), "%")))
		So("1.2e3.4", shouldTokenize, NewToken(NumberToken, floatNum("1.2e3")))
		So("-1e-1", shouldTokenize, NewToken(NumberToken, floatNum("-1e-1")))
		So("-1e+2em", shouldTokenize, NewToken(DimensionToken, dim(floatNum("-1e+2"), "em")))
		So(`1e\m`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "em")))
		So(`1\65\6d`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "em")))
		So(`1\000025`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "%")))
		So(`1\d888`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "\ufffd")))
		So(`1\110000`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "\ufffd")))
		So(`1\`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "\ufffd")))
		So("1e\\\n", shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "e")))
		So(`1test\`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "test\ufffd")))
		So("1-x", shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "-x")))
		So(`1-\`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "-\ufffd")))
		So(`1-\n`, shouldTokenize, NewToken(DimensionToken, dim(intNum("1"), "-n")))
		So("1-\\\n", shouldTokenize, NewToken(NumberToken, intNum("1")))
		So("1e900", shouldTokenize,
			NewErrorToken(&strconv.NumError{"ParseFloat", "1e900", strconv.ErrRange}))
	})

	Convey("identifier", t, func() {
		So("test", shouldTokenize, NewToken(IdentToken, "test"))
		So("test ing", shouldTokenize, NewToken(IdentToken, "test"))
		So("test(ing)", shouldTokenize, NewToken(FunctionToken, "test"))
		So(`\'test\'`, shouldTokenize, NewToken(IdentToken, "'test'"))
		So("\\\n", shouldTokenize, NewDelimToken('\\'))

		So("url(ing)", shouldTokenize, NewToken(UrlToken, "ing"))
		So("url(", shouldTokenize, NewToken(UrlToken, ""))
		So("url( ", shouldTokenize, NewToken(UrlToken, ""))
		So("url()", shouldTokenize, NewToken(UrlToken, ""))
		So("url(test)", shouldTokenize, NewToken(UrlToken, "test"))
		So("url(  test  )", shouldTokenize, NewToken(UrlToken, "test"))
		So("url('test')", shouldTokenize, NewToken(UrlToken, "test"))
		So("url(  'test'  )", shouldTokenize, NewToken(UrlToken, "test"))
		So(`url('test\'test')`, shouldTokenize, NewToken(UrlToken, "test'test"))
		So(`url(\'test\')`, shouldTokenize, NewToken(UrlToken, "'test'"))
		So("url('", shouldTokenize, NewToken(UrlToken, ""))
		So("url(''   ,", shouldTokenize, NewToken(BadUrlToken, nil))
		So("url((", shouldTokenize, NewToken(BadUrlToken, nil))
		So("url(x'", shouldTokenize, NewToken(BadUrlToken, nil))
		So("url(x\\\n", shouldTokenize, NewToken(BadUrlToken, nil))
		So("url(x \nx", shouldTokenize, NewToken(BadUrlToken, nil))
		So("url('x\n')", shouldTokenize, NewToken(BadUrlToken, nil))
		So("url('x\n\\)x)y", shouldTokenize, NewToken(BadUrlToken, nil), NewToken(IdentToken, "y"))
		So("url(\001)", shouldTokenize, NewToken(BadUrlToken, nil))
		So("url(\\001)", shouldTokenize, NewToken(UrlToken, "\001"))
	})

	Convey("unicode range", t, func() {
		urange := func(start, end rune) *Token {
			return NewToken(UnicodeRangeToken, UnicodeRange{start, end})
		}

		So("u+u", shouldTokenize, NewToken(IdentToken, "u"))
		So("u+0", shouldTokenize, urange(0, 0))
		So("u+00100?", shouldTokenize, urange(0x1000, 0x100f))
		So("u+001???", shouldTokenize, urange(0x1000, 0x1fff))
		So("u+001000?", shouldTokenize, urange(0x1000, 0x1000))
		So("u+1000-1011", shouldTokenize, urange(0x1000, 0x1011))
		So("u+1000-101?", shouldTokenize, urange(0x1000, 0x101))
		So("u+100?-1011", shouldTokenize, urange(0x1000, 0x100f))
	})

	Convey("EOF", t, func() {
		So("", shouldTokenize, NewEOFToken())
	})

	Convey("anything else", t, func() {
		So("!", shouldTokenize, NewDelimToken('!'))
	})
}
