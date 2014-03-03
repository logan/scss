package css3

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type TokenType int

const (
	ErrorToken TokenType = iota

	IdentToken
	FunctionToken
	AtKeywordToken
	HashToken
	StringToken
	BadStringToken
	UrlToken
	BadUrlToken
	DelimToken
	NumberToken
	PercentageToken
	DimensionToken
	UnicodeRangeToken
	IncludeMatchToken
	DashMatchToken
	PrefixMatchToken
	SuffixMatchToken
	SubstringMatchToken
	ColumnToken
	WhitespaceToken
	CDOToken
	CDCToken
	ColonToken
	SemicolonToken
	CommaToken
	LParenToken
	RParenToken
	LSquareToken
	RSquareToken
	LCurlyToken
	RCurlyToken
	EOFToken

	MinTokenType = IdentToken
	MaxTokenType = EOFToken
)

func (tt TokenType) String() string {
	switch tt {
	case ErrorToken:
		return "ErrorToken"
	case IdentToken:
		return "IdentToken"
	case FunctionToken:
		return "FunctionToken"
	case AtKeywordToken:
		return "AtKeywordToken"
	case HashToken:
		return "HashToken"
	case StringToken:
		return "StringToken"
	case BadStringToken:
		return "BadStringToken"
	case UrlToken:
		return "UrlToken"
	case BadUrlToken:
		return "BadUrlToken"
	case DelimToken:
		return "DelimToken"
	case NumberToken:
		return "NumberToken"
	case PercentageToken:
		return "PercentageToken"
	case DimensionToken:
		return "DimensionToken"
	case UnicodeRangeToken:
		return "UnicodeRangeToken"
	case IncludeMatchToken:
		return "IncludeMatchToken"
	case DashMatchToken:
		return "DashMatchToken"
	case PrefixMatchToken:
		return "PrefixMatchToken"
	case SuffixMatchToken:
		return "SuffixMatchToken"
	case SubstringMatchToken:
		return "SubstringMatchToken"
	case ColumnToken:
		return "ColumnToken"
	case WhitespaceToken:
		return "WhitespaceToken"
	case CDOToken:
		return "CDOToken"
	case CDCToken:
		return "CDCToken"
	case ColonToken:
		return "ColonToken"
	case SemicolonToken:
		return "SemicolonToken"
	case CommaToken:
		return "CommaToken"
	case LParenToken:
		return "LParenToken"
	case RParenToken:
		return "RParenToken"
	case LSquareToken:
		return "LSquareToken"
	case RSquareToken:
		return "RSquareToken"
	case LCurlyToken:
		return "LCurlyToken"
	case RCurlyToken:
		return "RCurlyToken"
	case EOFToken:
		return "EOFToken"
	default:
		return "UnknownToken"
	}
}

type Token struct {
	TokenType
	Value interface{}
}

func NewToken(tokenType TokenType, value interface{}) *Token { return &Token{tokenType, value} }
func NewEOFToken() *Token                                    { return NewToken(EOFToken, nil) }
func NewDelimToken(ch rune) *Token                           { return NewToken(DelimToken, ch) }
func NewErrorToken(err error) *Token                         { return NewToken(ErrorToken, err) }

func (t Token) String() string {
	return fmt.Sprintf("{%s=%v}", t.TokenType, t.Value)
}

type Identifier string
type NumberType int

const (
	Integer = iota
	Float
)

type Numeric struct {
	NumberType
	Repr    string
	Integer int64
	Float   float64
	Unit    string
}

func (n Numeric) String() string {
	t := "int"
	if n.NumberType == Float {
		t = "float"
	}
	return fmt.Sprintf("<num(%s):%s%s>", t, n.Repr, n.Unit)
}

func (n *Numeric) Float64() float64 {
	switch n.NumberType {
	case Integer:
		return float64(n.Integer)
	case Float:
		return n.Float
	default:
		return 0
	}
}

func (n *Numeric) parse(repr string) (err error) {
	n.Repr = repr
	switch n.NumberType {
	case Integer:
		n.Integer, err = strconv.ParseInt(n.Repr, 0, 64)
	case Float:
		n.Float, err = strconv.ParseFloat(n.Repr, 64)
	}
	return
}

type UnicodeRange struct {
	Start rune
	End   rune
}

type Tokenizer struct {
	*Scanner
}

func NewTokenizer(runeScanner io.RuneScanner) *Tokenizer {
	return &Tokenizer{NewScanner(runeScanner)}
}

func (tk *Tokenizer) ConsumeToken() *Token {
	var ch rune
	for tk.Error() == nil {
		ch = tk.Consume1()
		if ch == EOFRune {
			return NewEOFToken()
		}
		if isWhitespace(ch) {
			tk.skipWhitespace()
			tk.Reconsume()
			return NewToken(WhitespaceToken, nil)
		}
		if ch != '/' {
			break
		}
		if tk.Next() == '*' {
			star := false
			tk.Consume1()
			for tk.Error() == nil && tk.Current() != EOFRune {
				tk.Consume1()
				if star {
					if tk.Current() == '/' {
						break
					}
					star = false
				}
				star = tk.Current() == '*'
			}
		} else {
			return NewDelimToken(ch)
		}
	}
	if tk.Error() != nil {
		return NewErrorToken(tk.Error())
	}
	switch ch {
	case '"', '\'':
		return tk.consumeStringToken(ch)
	case '#':
		next3 := tk.Peek3()
		if isName(next3[0]) || (next3[0] == '\\' && next3[1] != '\n') {
			isIdent := startsIdent(tk.Peek3())
			tk.Consume1()
			name := tk.consumeName()
			if isIdent {
				return NewToken(HashToken, Identifier(name))
			}
			return NewToken(HashToken, name)
		}
		return NewDelimToken(ch)
	case ',':
		return NewToken(CommaToken, nil)
	case ':':
		return NewToken(ColonToken, nil)
	case ';':
		return NewToken(SemicolonToken, nil)
	case '<':
		if tk.PeekString() == "!--" {
			tk.Consume(3)
			return NewToken(CDOToken, nil)
		}
		return NewDelimToken(ch)
	case '@':
		if startsIdent(tk.Peek3()) {
			tk.Consume(1)
			name := tk.consumeName()
			return NewToken(AtKeywordToken, name)
		}
		return NewDelimToken(ch)
	case '\\':
		if tk.Next() != '\n' {
			return tk.consumeIdentLike()
		}
		// Technically this is a parse error.
		return NewDelimToken(ch)
	case '(':
		return NewToken(LParenToken, nil)
	case ')':
		return NewToken(RParenToken, nil)
	case '[':
		return NewToken(LSquareToken, nil)
	case ']':
		return NewToken(RSquareToken, nil)
	case '{':
		return NewToken(LCurlyToken, nil)
	case '}':
		return NewToken(RCurlyToken, nil)
	case '$':
		return tk.delimOrMatchToken(ch, SuffixMatchToken)
	case '*':
		return tk.delimOrMatchToken(ch, SubstringMatchToken)
	case '+':
		next3 := tk.Peek3()
		if isDigit(next3[0]) || (next3[0] == '.' && isDigit(next3[1])) {
			return tk.consumeNumeric()
		}
		return NewDelimToken(ch)
	case '-':
		next3 := tk.Peek3()
		if isDigit(next3[0]) || (next3[0] == '.' && isDigit(next3[1])) {
			return tk.consumeNumeric()
		}
		if startsIdent([]rune{tk.Current(), next3[0], next3[1]}) {
			return tk.consumeIdentLike()
		}
		if next3[0] == '-' && next3[1] == '>' {
			tk.Consume(2)
			return NewToken(CDCToken, nil)
		}
		return NewDelimToken(ch)
	case '.':
		if isDigit(tk.Next()) {
			return tk.consumeNumeric()
		}
		return NewDelimToken(ch)
	case '^':
		return tk.delimOrMatchToken(ch, PrefixMatchToken)
	case '|':
		switch tk.Next() {
		case '=':
			tk.Consume1()
			return NewToken(DashMatchToken, nil)
		case '|':
			tk.Consume1()
			return NewToken(ColumnToken, nil)
		default:
			return NewDelimToken(ch)
		}
	case '~':
		return tk.delimOrMatchToken(ch, IncludeMatchToken)
	case 'U', 'u':
		next3 := tk.Peek3()
		if next3[0] == '+' && (next3[1] == '?' || isHexDigit(next3[1])) {
			return tk.consumeUnicodeRange()
		} else {
			return tk.consumeIdentLike()
		}
	default:
		if isDigit(ch) {
			return tk.consumeNumeric()
		}
		if isNameStart(ch) {
			return tk.consumeIdentLike()
		}
		return NewDelimToken(ch)
	}
}

func (tk *Tokenizer) consumeNumeric() *Token {
	num, err := tk.consumeNumber()
	if err != nil {
		return NewErrorToken(err)
	}
	next3 := tk.Peek3()
	if startsIdent([]rune{tk.Current(), next3[0], next3[1]}) {
		name := tk.consumeName()
		num.Unit = name
		return NewToken(DimensionToken, num)
	}
	if tk.Current() == '%' {
		num.Unit = "%"
		return NewToken(PercentageToken, num)
	}
	tk.Reconsume()
	return NewToken(NumberToken, num)
}

func (tk *Tokenizer) consumeNumber() (*Numeric, error) {
	num := new(Numeric)
	repr := bytes.NewBuffer(make([]byte, 0, 8))
	cur := tk.Current()

	if cur == '+' || cur == '-' {
		repr.WriteRune(cur)
		cur = tk.Consume1()
	}

	for isDigit(cur) {
		repr.WriteRune(cur)
		cur = tk.Consume1()
	}

	if cur == '.' && isDigit(tk.Next()) {
		num.NumberType = Float
		repr.WriteRune(cur)
		cur = tk.Consume1()
		for isDigit(cur) {
			repr.WriteRune(tk.Current())
			cur = tk.Consume1()
		}
	}

	next3 := tk.Peek3()
	if (cur == 'e' || cur == 'E') && (isDigit(next3[0]) || ((next3[0] == '+' || next3[0] == '-') && isDigit(next3[1]))) {
		num.NumberType = Float
		repr.WriteRune(cur)
		cur = tk.Consume1()
		if cur == '+' || cur == '-' {
			repr.WriteRune(cur)
			cur = tk.Consume1()
		}
		for isDigit(cur) {
			repr.WriteRune(cur)
			cur = tk.Consume1()
		}
	}

	return num, num.parse(repr.String())
}

func (tk *Tokenizer) consumeName() string {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	ch := tk.Current()
	ch2 := tk.Next()
	tk.Reconsume()
	for {
		if ch == '-' || isNameStart(ch) || isDigit(ch) {
			buf.WriteRune(ch)
			tk.Consume1()
		} else if ch == '\\' && ch2 != '\n' {
			tk.Consume(1)
			buf.WriteRune(tk.consumeEscape())
		} else {
			break
		}
		next3 := tk.Peek3()
		ch = next3[0]
		ch2 = next3[1]
	}
	return buf.String()
}

func (tk *Tokenizer) consumeIdentLike() *Token {
	name := tk.consumeName()
	if tk.Next() == '(' {
		if strings.ToLower(name) == "url" {
			return tk.consumeUrl()
		}
		tk.Consume1()
		return NewToken(FunctionToken, name)
	}
	return NewToken(IdentToken, Identifier(name))
}

func (tk *Tokenizer) consumeUrl() *Token {
	tok := NewToken(UrlToken, "")
	tk.Consume(2)
	tk.skipWhitespace()
	cur := tk.Current()
	if cur < 0 {
		return tok
	}
	if cur == '"' || cur == '\'' {
		strTok := tk.consumeStringToken(cur)
		if strTok.TokenType == BadStringToken {
			return tk.consumeBadUrlRemnants()
		}
		tk.Consume1()
		tk.skipWhitespace()
		if tk.Current() != ')' && tk.Current() != EOFRune {
			return tk.consumeBadUrlRemnants()
		}
		tok.Value = strTok.Value
		return tok
	}

	buf := bytes.NewBuffer(make([]byte, 0, 64))
	for {
		if cur < 0 || cur == ')' {
			break
		}
		if isWhitespace(cur) {
			tk.skipWhitespace()
			cur = tk.Current()
			if cur != ')' {
				return tk.consumeBadUrlRemnants()
			}
			break
		}
		if cur == '"' || cur == '\'' || cur == '(' || isNonPrintable(cur) {
			return tk.consumeBadUrlRemnants()
		}
		if cur == '\\' {
			if tk.Next() == '\n' {
				return tk.consumeBadUrlRemnants()
			}
			cur = tk.consumeEscape()
		}
		buf.WriteRune(cur)
		cur = tk.Consume1()
	}
	tok.Value = buf.String()
	return tok
}

func (tk *Tokenizer) consumeBadUrlRemnants() *Token {
	for {
		if tk.Current() == EOFRune || tk.Current() == ')' {
			break
		}
		if tk.Current() == '\\' && tk.Next() != '\n' {
			tk.consumeEscape()
		}
		tk.Consume1()
	}
	return NewToken(BadUrlToken, nil)
}

func (tk *Tokenizer) delimOrMatchToken(ch rune, matchType TokenType) *Token {
	if tk.Next() == '=' {
		tk.Consume1()
		return NewToken(matchType, nil)
	}
	return NewDelimToken(ch)
}

func (tk *Tokenizer) consumeStringToken(delim rune) *Token {
	tt := StringToken
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	for {
		ch := tk.Consume1()
		if ch == EOFRune || ch == delim {
			break
		}
		if ch == '\n' {
			tk.Reconsume()
			tt = BadStringToken
			break
		}
		if ch == '\\' {
			if tk.Next() == EOFRune {
				break
			}
			if tk.Next() == '\n' {
				tk.Consume1()
				continue
			}
			ch = tk.consumeEscape()
		}
		buf.WriteRune(ch)
	}
	return NewToken(tt, buf.String())
}

func (tk *Tokenizer) consumeHexCode(max int) (code int, length int) {
	tk.Reconsume()
	for length < max {
		code = code*16 + parseHexDigit(tk.Consume1())
		length++
		if !isHexDigit(tk.Next()) {
			break
		}
	}
	return
}

func (tk *Tokenizer) consumeEscape() rune {
	cur := tk.Consume1()
	if cur == EOFRune {
		return '\ufffd'
	}
	if !isHexDigit(cur) {
		return cur
	}
	code, _ := tk.consumeHexCode(6)
	if isWhitespace(tk.Next()) {
		tk.Consume1()
	}
	if code == 0 || (code >= 0xd800 && code <= 0xdfff) || code >= 0x10ffff {
		code = 0xfffd
	}
	return rune(code)
}

func (tk *Tokenizer) consumeUnicodeRange() *Token {
	var code, length int
	tk.Consume(2) // skip leading + and consume first digit
	if isHexDigit(tk.Current()) {
		code, length = tk.consumeHexCode(6)
		tk.Consume1()
	}
	var qs uint
	for ; qs < uint(6-length) && tk.Current() == '?'; qs++ {
		tk.Consume1()
	}
	if qs > 0 {
		start := rune(code << (4 * qs))
		end := rune(start | ((1 << (4 * qs)) - 1))
		tk.Reconsume()
		return NewToken(UnicodeRangeToken, UnicodeRange{start, end})
	}
	start := rune(code)
	end := start
	if tk.Current() == '-' && isHexDigit(tk.Next()) {
		tk.Consume(1) // skip leading -
		code, _ := tk.consumeHexCode(6)
		end = rune(code)
	} else {
		tk.Reconsume()
	}
	return NewToken(UnicodeRangeToken, UnicodeRange{start, end})
}

func (tk *Tokenizer) skipWhitespace() {
	for isWhitespace(tk.Current()) {
		tk.Consume1()
	}
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return parseHexDigit(ch) >= 0
}

func parseHexDigit(ch rune) int {
	idx := strings.IndexRune("0123456789abcdefABCDEF", ch)
	if idx >= 16 {
		idx -= 6
	}
	return idx
}

func isWhitespace(ch rune) bool {
	return strings.IndexRune(" \n\t", ch) >= 0
}

func startsIdent(next3 []rune) bool {
	if next3[0] == '-' {
		next3 = next3[1:]
	}
	if isNameStart(next3[0]) {
		return true
	}
	if startsEscape(next3) {
		return true
	}
	return false
}

func isNameStart(ch rune) bool {
	if ch == '_' || ch >= 0x0080 || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
		return true
	}
	return false
}

func isName(ch rune) bool {
	return ch == '-' || isNameStart(ch) || isDigit(ch)
}

func startsEscape(next2 []rune) bool {
	return next2[0] == '\\' && next2[1] != '\n'
}

func isNonPrintable(ch rune) bool {
	if ch == 0xb || ch == 0x7f || (ch >= 0 && ch <= 8) || (ch >= 0xe && ch <= 0x1f) {
		return true
	}
	return false
}
