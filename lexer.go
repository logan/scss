package sass

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenType int

const (
	tokenError tokenType = iota

	tokenPeriod
	tokenComma
	tokenColon
	tokenSemicolon
	tokenLeftBrace
	tokenRightBrace
	tokenLeftBracket
	tokenRightBracket
	tokenLeftParen
	tokenRightParen
	tokenAmp
	tokenPlus
	tokenMinus
	tokenSlash
	tokenStar
	tokenAnd
	tokenOr
	tokenEq
	tokenNe
	tokenLte
	tokenLt
	tokenGte
	tokenGt
	tokenImportant
	tokenDefault
	tokenDebug
	tokenWarn
	tokenInclude
	tokenExtend
	tokenIf
	tokenElse
	tokenElseIf
	tokenFor
	tokenMixin
	tokenFunction
	tokenReturn
	tokenOption
	tokenImport
	tokenMedia
	tokenFontFace
	tokenVariables
	tokenPage
	tokenCharset

	tokenComment
	tokenIdent
	tokenRGB
	tokenVar
	tokenNumber

	tokenEof
)

type token struct {
	tokenType
	offset int
	width  int
	value  interface{}
}

type unit int

const (
	none unit = iota
	percent

	em
	ex
	ch
	rem
	vw
	vh
	vmin
	vmax
	cm
	mm
	in
	px
	pt
	pc
)

type number struct {
	float64
	unit
}

type lexer struct {
	*File
	offset    int
	stack     []int
	lastWidth int
	emitter   chan token
}

type color struct {
	r byte
	g byte
	b byte
	a byte
}

func (c *color) parse(s string) error {
	if s[0] != '#' || (len(s) != 4 && len(s) != 7) {
		return errors.New("invalid rgb")
	}
	if len(s) == 4 {
		v, _ := strconv.ParseInt(s[1:], 16, 16)
		c.r = byte((v & 0xf00) >> 8)
		c.g = byte((v & 0x0f0) >> 4)
		c.b = byte(v & 0x00f)
		c.a = 0
		c.r |= c.r << 4
		c.g |= c.g << 4
		c.b |= c.b << 4
		return nil
	}
	v, _ := strconv.ParseInt(s[1:], 16, 32)
	c.r = byte((v & 0xff0000) >> 16)
	c.g = byte((v & 0x00ff00) >> 8)
	c.b = byte(v & 0x0000ff)
	c.a = 0
	return nil
}

func lexFile(f *File) chan token {
	lexer := &lexer{File: f, emitter: make(chan token, 1024)}
	go lexer.run()
	return lexer.emitter
}

func (lex *lexer) run() {
	for lex.offset < len(lex.File.Bytes) {
		if lex.space() {
			continue
		}
		t := token{offset: lex.offset}
		if lex.comment(&t) {
			lex.emit(t)
		} else if lex.variable(&t) {
			lex.emit(t)
		} else if lex.rgb(&t) {
			lex.emit(t)
		} else if lex.number(&t) {
			lex.emit(t)
		} else if lex.keyword(&t) {
			lex.emit(t)
		} else if lex.ident(&t) {
			lex.emit(t)
		} else if lex.simpleToken(&t) {
			lex.emit(t)
		}
		if t.tokenType == tokenError {
			break
		}
	}
	lex.emit(token{tokenType: tokenEof, offset: lex.offset})
	close(lex.emitter)
}

func (lex *lexer) next() (ch rune, width int) {
	ch, width = firstRune(lex.File.Bytes[lex.offset:])
	lex.lastWidth = width
	lex.offset += width
	return ch, width
}

func (lex *lexer) unread() {
	lex.offset -= lex.lastWidth
	lex.lastWidth = 0
}

func (lex *lexer) peek() (ch rune, width int) {
	ch, width = lex.next()
	lex.unread()
	return ch, width
}

func (lex *lexer) emit(t token) {
	lex.emitter <- t
}

func (lex *lexer) space() (b bool) {
	for {
		ch, width := lex.next()
		if width == 0 || !unicode.IsSpace(ch) {
			lex.unread()
			return
		}
		b = true
	}
}

func (lex *lexer) keyword(t *token) bool {
	nextBytes := lex.File.Bytes[lex.offset:]
	var tt int
	tt, t.value, t.width = keywordTrie.longestPrefix(nextBytes)
	t.tokenType = tokenType(tt)
	lex.offset += t.width
	return t.width > 0
}

func (lex *lexer) simpleToken(t *token) bool {
	nextBytes := lex.File.Bytes[lex.offset:]
	var tt int
	tt, t.value, t.width = simpleTokenTrie.longestPrefix(nextBytes)
	t.tokenType = tokenType(tt)
	lex.offset += t.width
	return t.width > 0
}

func (lex *lexer) comment(t *token) bool {
	prevOffset := lex.offset

	ch, width := lex.next()
	if ch != '/' {
		lex.unread()
		return false
	}

	ch, width = lex.next()
	switch ch {
	case '/':
		begin := lex.offset
		ch, width = lex.next()
		for width > 0 && ch != '\n' {
			ch, width = lex.next()
		}
		lex.unread()
		t.tokenType = tokenComment
		t.value = string(lex.File.Bytes[begin:lex.offset])
		t.width = lex.offset - prevOffset
	case '*':
		var starOffset int
		begin := lex.offset
		star := false
		ch, width = lex.next()
		for !star || ch != '/' {
			if width == 0 {
				t.value = errors.New("unterminated comment")
				return true
			}
			star = ch == '*'
			if star {
				starOffset = lex.offset - width
			}
			ch, width = lex.next()
		}
		t.tokenType = tokenComment
		t.value = string(lex.File.Bytes[begin:starOffset])
		t.width = lex.offset - prevOffset
	default:
		lex.unread()
		lex.offset = prevOffset
		return false
	}
	return true
}

func isHexDigit(ch rune) bool {
	return strings.IndexRune("0123456789abcdefABCDEF", ch) >= 0
}

func (lex *lexer) rgb(t *token) bool {
	begin := lex.offset
	ch, width := lex.next()
	if ch != '#' {
		lex.unread()
		return false
	}
	ch, width = lex.next()
	for width > 0 && isHexDigit(ch) {
		ch, width = lex.next()
	}
	lex.unread()

	var c color
	if err := c.parse(string(lex.File.Bytes[begin:lex.offset])); err != nil {
		t.value = err
		return true
	}
	t.tokenType = tokenRGB
	t.value = c
	t.width = lex.offset - begin
	return true
}

func (lex *lexer) ident(t *token) bool {
	begin := lex.offset
	ch, width := lex.next()
	if ch == '-' {
		ch, width = lex.peek()
		if ch != '_' && !unicode.IsLetter(ch) {
			lex.offset = begin
			return false
		}
	} else if ch != '_' && !unicode.IsLetter(ch) {
		lex.unread()
		return false
	}
	for width > 0 && (ch == '_' || ch == '-' || unicode.IsLetter(ch) || unicode.IsDigit(ch)) {
		ch, width = lex.next()
	}
	lex.unread()
	t.tokenType = tokenIdent
	t.value = string(lex.File.Bytes[begin:lex.offset])
	t.width = lex.offset - begin
	return true
}

func (lex *lexer) variable(t *token) bool {
	ch, width := lex.next()
	if ch != '$' {
		lex.unread()
		return false
	}
	if !lex.ident(t) {
		t.value = errors.New("invalid variable")
		return true
	}
	t.tokenType = tokenVar
	t.width += width
	return true
}

func (lex *lexer) number(t *token) bool {
	begin := lex.offset
	ch, width := lex.next()
	if ch == '-' || ch == '+' {
		ch, width = lex.next()
		if ch != '.' && !unicode.IsDigit(ch) {
			lex.offset = begin
			return false
		}
	} else if ch == '.' {
		ch, width = lex.peek()
		if !unicode.IsDigit(ch) {
			lex.offset = begin
			return false
		}
	} else if !unicode.IsDigit(ch) {
		lex.offset = begin
		return false
	}
	for width > 0 && unicode.IsDigit(ch) {
		ch, width = lex.next()
	}
	if ch == '.' {
		ch, width = lex.next()
		for width > 0 && unicode.IsDigit(ch) {
			ch, width = lex.next()
		}
	}
	if ch == 'e' {
		o := lex.offset - width
		ch, width = lex.next()
		if ch == '-' || ch == '+' || unicode.IsDigit(ch) {
			for width > 0 && unicode.IsDigit(ch) {
				ch, width = lex.next()
			}
			lex.unread()
		} else {
			lex.unread()
			lex.offset = o
		}
	} else {
		lex.unread()
	}

	var num number
	num.float64, t.value = strconv.ParseFloat(string(lex.File.Bytes[begin:lex.offset]), 64)
	if t.value != nil {
		return true
	}
	t.tokenType = tokenNumber

	u, _, uWidth := unitTrie.longestPrefix(lex.File.Bytes[lex.offset:])
	num.unit = unit(u)
	lex.offset += uWidth
	t.value = num
	t.width = lex.offset - begin
	return true
}

type trie struct {
	terminal int
	children map[rune]*trie
}

func newTrie() *trie {
	return &trie{children: make(map[rune]*trie)}
}

func newTrieFrom(tokens map[string]int) *trie {
	t := newTrie()
	t.addAll(tokens)
	return t
}

func (t *trie) add(s string, tt int) {
	for _, r := range s {
		next, ok := t.children[r]
		if !ok {
			next = newTrie()
			t.children[r] = next
		}
		t = next
	}
	t.terminal = tt
}

func (t *trie) addAll(tokens map[string]int) {
	for k, v := range tokens {
		t.add(k, v)
	}
}

func (t *trie) longestPrefix(b []byte) (tt int, value string, width int) {
	depth := 0
	for {
		ch, size := firstRune(b[depth:])
		if size == 0 {
			value = string(b[:width])
			return
		}
		next, ok := t.children[ch]
		if !ok {
			value = string(b[:width])
			return
		}
		depth += size
		if next.terminal > 0 {
			tt = next.terminal
			width = depth
		}
		t = next
	}
}

func firstRune(b []byte) (ch rune, width int) {
	if len(b) == 0 {
		return 0, 0
	}
	if b[0] < utf8.RuneSelf {
		return rune(b[0]), 1
	}
	return utf8.DecodeRune(b)
}

var keywords = map[string]int{
	"and": int(tokenAnd),
	"or":  int(tokenOr),
}

var keywordTrie = newTrieFrom(keywords)

var simpleTokens = map[string]int{
	".":          int(tokenPeriod),
	",":          int(tokenComma),
	":":          int(tokenColon),
	";":          int(tokenSemicolon),
	"{":          int(tokenLeftBrace),
	"}":          int(tokenRightBrace),
	"[":          int(tokenLeftBracket),
	"]":          int(tokenRightBracket),
	"(":          int(tokenLeftParen),
	")":          int(tokenRightParen),
	"&":          int(tokenAmp),
	"+":          int(tokenPlus),
	"-":          int(tokenMinus),
	"/":          int(tokenSlash),
	"*":          int(tokenStar),
	"==":         int(tokenEq),
	"!=":         int(tokenNe),
	"<=":         int(tokenLte),
	"<":          int(tokenLt),
	">=":         int(tokenGte),
	">":          int(tokenGt),
	"!important": int(tokenImportant),
	"!default":   int(tokenDefault),
	"@debug":     int(tokenDebug),
	"@warn":      int(tokenWarn),
	"@include":   int(tokenInclude),
	"@extend":    int(tokenExtend),
	"@if":        int(tokenIf),
	"@else":      int(tokenElse),
	"@else if":   int(tokenElseIf),
	"@for":       int(tokenFor),
	"@mixin":     int(tokenMixin),
	"@function":  int(tokenFunction),
	"@return":    int(tokenReturn),
	"@option":    int(tokenOption),
	"@import":    int(tokenImport),
	"@media":     int(tokenMedia),
	"@font-face": int(tokenFontFace),
	"@variables": int(tokenVariables),
	"@vars":      int(tokenVariables),
	"@page":      int(tokenPage),
	"@charset":   int(tokenCharset),
}

var simpleTokenTrie = newTrieFrom(simpleTokens)

var units = map[string]int{
	"":     int(none),
	"%":    int(percent),
	"em":   int(em),
	"ex":   int(ex),
	"ch":   int(ch),
	"rem":  int(rem),
	"vw":   int(vw),
	"vh":   int(vh),
	"vmin": int(vmin),
	"vmax": int(vmax),
	"cm":   int(cm),
	"mm":   int(mm),
	"in":   int(in),
	"px":   int(px),
	"pt":   int(pt),
	"pc":   int(pc),
}

var unitTrie = newTrieFrom(units)
