package sass

import (
	"errors"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func testFile(contents string) *File {
	f, err := openFile(nil, &stringFile{contents: []byte(contents)}, "")
	if err != nil {
		panic(err)
	}
	return f
}

func TestSimpleTokens(t *testing.T) {
	tokensSeen := make([]tokenType, 0)
	unitsSeen := make([]unit, 0)
	tokenize := func(contents string) []token {
		tokens := make([]token, 0)
		for t := range lexFile(testFile(contents)) {
			tokensSeen = append(tokensSeen, t.tokenType)
			if num, ok := t.value.(number); ok {
				unitsSeen = append(unitsSeen, num.unit)
			}
			tokens = append(tokens, t)
		}
		return tokens
	}

	Convey("simple tokenization", t, func() {
		tokens := tokenize(".,:;{}[]()&+-*")
		expected := []token{
			{tokenPeriod, 0, 1, "."},
			{tokenComma, 1, 1, ","},
			{tokenColon, 2, 1, ":"},
			{tokenSemicolon, 3, 1, ";"},
			{tokenLeftBrace, 4, 1, "{"},
			{tokenRightBrace, 5, 1, "}"},
			{tokenLeftBracket, 6, 1, "["},
			{tokenRightBracket, 7, 1, "]"},
			{tokenLeftParen, 8, 1, "("},
			{tokenRightParen, 9, 1, ")"},
			{tokenAmp, 10, 1, "&"},
			{tokenPlus, 11, 1, "+"},
			{tokenMinus, 12, 1, "-"},
			{tokenStar, 13, 1, "*"},
			{tokenEof, 14, 0, nil},
		}
		So(tokens, ShouldResemble, expected)

		tokens = tokenize("andor==!=<=<>=>")
		expected = []token{
			{tokenAnd, 0, 3, "and"},
			{tokenOr, 3, 2, "or"},
			{tokenEq, 5, 2, "=="},
			{tokenNe, 7, 2, "!="},
			{tokenLte, 9, 2, "<="},
			{tokenLt, 11, 1, "<"},
			{tokenGte, 12, 2, ">="},
			{tokenGt, 14, 1, ">"},
			{tokenEof, 15, 0, nil},
		}
		So(tokens, ShouldResemble, expected)

		tokens = tokenize("!important !default @debug @warn @include @extend")
		expected = []token{
			{tokenImportant, 0, 10, "!important"},
			{tokenDefault, 11, 8, "!default"},
			{tokenDebug, 20, 6, "@debug"},
			{tokenWarn, 27, 5, "@warn"},
			{tokenInclude, 33, 8, "@include"},
			{tokenExtend, 42, 7, "@extend"},
			{tokenEof, 49, 0, nil},
		}
		So(tokens, ShouldResemble, expected)

		tokens = tokenize("@if @else @else if @for @mixin @function @return")
		expected = []token{
			{tokenIf, 0, 3, "@if"},
			{tokenElse, 4, 5, "@else"},
			{tokenElseIf, 10, 8, "@else if"},
			{tokenFor, 19, 4, "@for"},
			{tokenMixin, 24, 6, "@mixin"},
			{tokenFunction, 31, 9, "@function"},
			{tokenReturn, 41, 7, "@return"},
			{tokenEof, 48, 0, nil},
		}
		So(tokens, ShouldResemble, expected)

		tokens = tokenize("@option @import @media @font-face @variables @vars @page @charset")
		expected = []token{
			{tokenOption, 0, 7, "@option"},
			{tokenImport, 8, 7, "@import"},
			{tokenMedia, 16, 6, "@media"},
			{tokenFontFace, 23, 10, "@font-face"},
			{tokenVariables, 34, 10, "@variables"},
			{tokenVariables, 45, 5, "@vars"},
			{tokenPage, 51, 5, "@page"},
			{tokenCharset, 57, 8, "@charset"},
			{tokenEof, 65, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("skipping whitespace", t, func() {
		tokens := tokenize(". ,  : \n ;  ")
		expected := []token{
			{tokenPeriod, 0, 1, "."},
			{tokenComma, 2, 1, ","},
			{tokenColon, 5, 1, ":"},
			{tokenSemicolon, 9, 1, ";"},
			{tokenEof, 12, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("comments", t, func() {
		tokens := tokenize("/ //* \n*/ /*\ntest */")
		expected := []token{
			{tokenSlash, 0, 1, "/"},
			{tokenComment, 2, 4, "* "},
			{tokenStar, 7, 1, "*"},
			{tokenSlash, 8, 1, "/"},
			{tokenComment, 10, 10, "\ntest "},
			{tokenEof, 20, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("numbers", t, func() {
		tokens := tokenize("0 1 -1 +1 0.1 1. -0.1 +.1")
		expected := []token{
			{tokenNumber, 0, 1, number{0, none}},
			{tokenNumber, 2, 1, number{1, none}},
			{tokenNumber, 4, 2, number{-1, none}},
			{tokenNumber, 7, 2, number{1, none}},
			{tokenNumber, 10, 3, number{0.1, none}},
			{tokenNumber, 14, 2, number{1, none}},
			{tokenNumber, 17, 4, number{-0.1, none}},
			{tokenNumber, 22, 3, number{0.1, none}},
			{tokenEof, 25, 0, nil},
		}
		So(tokens, ShouldResemble, expected)

		tokens = tokenize("1ex 1e-x")
		expected = []token{
			{tokenNumber, 0, 3, number{1, ex}},
			{tokenError, 4, 0, &strconv.NumError{"ParseFloat", "1e", strconv.ErrSyntax}},
			{tokenEof, 6, 0, nil},
		}
		So(tokens, ShouldResemble, expected)

		tokens = tokenize("1 1e2% 2ex 3em 4ch 5rem 6vw 7vh 8vmin 9vmax 0cm 1mm 2in 3px 4pt 5pc")
		expected = []token{
			{tokenNumber, 0, 1, number{1, none}},
			{tokenNumber, 2, 4, number{100, percent}},
			{tokenNumber, 7, 3, number{2, ex}},
			{tokenNumber, 11, 3, number{3, em}},
			{tokenNumber, 15, 3, number{4, ch}},
			{tokenNumber, 19, 4, number{5, rem}},
			{tokenNumber, 24, 3, number{6, vw}},
			{tokenNumber, 28, 3, number{7, vh}},
			{tokenNumber, 32, 5, number{8, vmin}},
			{tokenNumber, 38, 5, number{9, vmax}},
			{tokenNumber, 44, 3, number{0, cm}},
			{tokenNumber, 48, 3, number{1, mm}},
			{tokenNumber, 52, 3, number{2, in}},
			{tokenNumber, 56, 3, number{3, px}},
			{tokenNumber, 60, 3, number{4, pt}},
			{tokenNumber, 64, 3, number{5, pc}},
			{tokenEof, 67, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("identifiers", t, func() {
		tokens := tokenize("ab a-b -a-b -_a - 0a -1 ---a0 _0ಠ")
		expected := []token{
			{tokenIdent, 0, 2, "ab"},
			{tokenIdent, 3, 3, "a-b"},
			{tokenIdent, 7, 4, "-a-b"},
			{tokenIdent, 12, 3, "-_a"},
			{tokenMinus, 16, 1, "-"},
			{tokenNumber, 18, 1, number{0, none}},
			{tokenIdent, 19, 1, "a"},
			{tokenNumber, 21, 2, number{-1, none}},
			{tokenMinus, 24, 1, "-"},
			{tokenMinus, 25, 1, "-"},
			{tokenIdent, 26, 3, "-a0"},
			{tokenIdent, 30, 5, "_0ಠ"},
			{tokenEof, 35, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("variables", t, func() {
		tokens := tokenize("$test $-test $ ")
		expected := []token{
			{tokenVar, 0, 5, "test"},
			{tokenVar, 6, 6, "-test"},
			{tokenError, 13, 0, errors.New("invalid variable")},
			{tokenEof, 14, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("rgb", t, func() {
		tokens := tokenize("#123 #123456 #1")
		expected := []token{
			{tokenRGB, 0, 4, color{0x11, 0x22, 0x33, 0}},
			{tokenRGB, 5, 7, color{0x12, 0x34, 0x56, 0}},
			{tokenError, 13, 0, errors.New("invalid rgb")},
			{tokenEof, 15, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("errors", t, func() {
		tokens := tokenize("/*")
		expected := []token{
			{tokenError, 0, 0, errors.New("unterminated comment")},
			{tokenEof, 2, 0, nil},
		}
		So(tokens, ShouldResemble, expected)

		tokens = tokenize("1e400/*")
		expected = []token{
			{tokenError, 0, 0, &strconv.NumError{"ParseFloat", "1e400", strconv.ErrRange}},
			{tokenEof, 5, 0, nil},
		}
		So(tokens, ShouldResemble, expected)
	})

	Convey("all token types have been tested", t, func() {
		for tt := tokenError + 1; tt < tokenEof; tt++ {
			So(tokensSeen, ShouldContain, tt)
		}
	})

	Convey("all numeric units have been tested", t, func() {
		for u := none; u <= pc; u++ {
			So(unitsSeen, ShouldContain, u)
		}
	})
}
