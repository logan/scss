package css3

import (
	"errors"
	"fmt"
	"io"
)

var (
	SyntaxErr          = errors.New("invalid syntax")
	EmptyErr           = errors.New("empty")
	ExtraInputErr      = errors.New("extra input")
	UnmatchedCurlyErr  = errors.New("unexpected }")
	UnmatchedSquareErr = errors.New("unexpected ]")
	UnmatchedParenErr  = errors.New("unexpected )")
)

type Node interface {
	TestRepr() interface{}
}

func nodeIsTokenType(node Node, tokenType TokenType) bool {
	switch n := node.(type) {
	case EOFNode:
		return tokenType == EOFToken
	case *ErrorNode:
		return tokenType == ErrorToken
	case *TokenNode:
		return n.TokenType == tokenType
	default:
		return false
	}
}

func nodeIsEOFOrError(node Node) bool {
	switch node.(type) {
	case ErrorNode, EOFNode:
		return true
	default:
		return false
	}
}

func nodeListTestRepr(nl []Node) []interface{} {
	if nl == nil {
		return []interface{}{}
	}
	result := make([]interface{}, len(nl))
	for i, n := range nl {
		result[i] = n.TestRepr()
	}
	return result
}

func nodeListLastTwoNonwhitespace(nl []Node) (chopped []Node, penultimate Node, last Node) {
	var i int
	for i = len(nl) - 1; i >= 0 && (penultimate == nil || last == nil); i-- {
		switch x := nl[i].(type) {
		case *TokenNode:
			if x.TokenType == WhitespaceToken {
				continue
			}
		}
		if last == nil {
			last = nl[i]
		} else {
			penultimate = nl[i]
		}
	}
	if i >= 0 {
		return nl[:i+1], penultimate, last
	}
	return nl, nil, nil
}

type EOFNode int

func NewEOFNode() EOFNode { return EOFNode(0) }

func (n EOFNode) TestRepr() interface{} { return nil }

type ErrorNode struct{ error }

func NewErrorNode(err error) *ErrorNode { return &ErrorNode{err} }

func (n ErrorNode) TestRepr() interface{} {
	var repr string
	switch n.error {
	case SyntaxErr:
		repr = "invalid"
	case EmptyErr:
		repr = "empty"
	case ExtraInputErr:
		repr = "extra-input"
	case UnmatchedCurlyErr:
		repr = "}"
	case UnmatchedSquareErr:
		repr = "]"
	case UnmatchedParenErr:
		repr = ")"
	default:
		repr = "unknown"
	}
	return []interface{}{"error", repr}
}

type QualifiedRuleNode struct {
	Prelude []Node
	Body    []Node
}

func NewQualifiedRuleNode(prelude []Node, body []Node) *QualifiedRuleNode {
	return &QualifiedRuleNode{prelude, body}
}

func (n *QualifiedRuleNode) TestRepr() interface{} {
	return []interface{}{"qualified rule", nodeListTestRepr(n.Prelude), nodeListTestRepr(n.Body)}
}

type AtRuleNode struct {
	Name    string
	Prelude []Node
	Body    []Node
}

func NewAtRuleNode(name string, prelude []Node, body []Node) *AtRuleNode {
	return &AtRuleNode{name, prelude, body}
}

func (n *AtRuleNode) TestRepr() interface{} {
	var body interface{}
	if n.Body != nil {
		body = nodeListTestRepr(n.Body)
	}
	return []interface{}{"at-rule", n.Name, nodeListTestRepr(n.Prelude), body}
}

type BlockNode struct {
	EndDelim TokenType
	Values   []Node
}

func NewBlockNode(endDelim TokenType, values ...Node) *BlockNode {
	return &BlockNode{endDelim, values}
}

func (n *BlockNode) TestRepr() interface{} {
	var t string
	switch n.EndDelim {
	case RParenToken:
		t = "()"
	case RCurlyToken:
		t = "{}"
	case RSquareToken:
		t = "[]"
	}
	return append([]interface{}{t}, nodeListTestRepr(n.Values)...)
}

type FunctionNode struct {
	Name   string
	Values []Node
}

func NewFunctionNode(name string, values ...Node) *FunctionNode {
	return &FunctionNode{name, values}
}

func (n *FunctionNode) TestRepr() interface{} {
	return append([]interface{}{"function", n.Name}, nodeListTestRepr(n.Values)...)
}

func (n *FunctionNode) Params() []Node {
	var needComma bool
	params := make([]Node, 0)
	for _, v := range n.Values {
		if tn, ok := v.(*TokenNode); ok {
			switch tn.TokenType {
			case WhitespaceToken:
				continue
			case CommaToken:
				if needComma {
					needComma = false
					continue
				} else {
					return nil
				}
			}
		}
		if needComma {
			return nil
		}
		params = append(params, v)
		needComma = true
	}
	return params
}

func (n *FunctionNode) Color() *Color {
	name := toLower(n.Name)
	switch name {
	case "rgb":
		if cs, ok := n.requireFloats("", "", ""); ok {
			return RGB(cs[0]/255, cs[1]/255, cs[2]/255)
		}
		if cs, ok := n.requireFloats("%", "%", "%"); ok {
			return RGB(cs[0]/100, cs[1]/100, cs[2]/100)
		}
	case "rgba":
		if cs, ok := n.requireFloats("", "", "", ""); ok {
			return RGBA(cs[0]/255, cs[1]/255, cs[2]/255, cs[3])
		}
		if cs, ok := n.requireFloats("%", "%", "%", ""); ok {
			return RGBA(cs[0]/100, cs[1]/100, cs[2]/100, cs[3])
		}
	case "hsl":
		if cs, ok := n.requireFloats("", "%", "%"); ok {
			return HSL(normDeg(cs[0]), cs[1]/100, cs[2]/100)
		}
		return nil
	case "hsla":
		if cs, ok := n.requireFloats("", "%", "%", ""); ok {
			return HSLA(normDeg(cs[0]), cs[1]/100, cs[2]/100, cs[3])
		}
	}
	return nil
}

func (n *FunctionNode) requireFloats(units ...string) ([]float64, bool) {
	params := n.Params()
	if len(params) != len(units) {
		return nil, false
	}
	for i, unit := range units {
		if num, ok := params[i].(*NumberNode); !ok || num.Unit != unit {
			return nil, false
		}
	}
	result := make([]float64, len(params))
	for i, param := range params {
		num := param.(*NumberNode)
		if num.NumberType == Integer {
			result[i] = float64(num.Integer)
		} else {
			result[i] = num.Float
		}
	}
	return result, true
}

type HashNode struct {
	Hash         string
	Unrestricted bool
}

func NewHashNode(hash string, unrestricted bool) *HashNode { return &HashNode{hash, unrestricted} }

func (n *HashNode) TestRepr() interface{} {
	var rest string
	if n.Unrestricted {
		rest = "unrestricted"
	} else {
		rest = "id"
	}
	return []interface{}{"hash", n.Hash, rest}
}

type NumberNode struct {
	*Numeric
	Type string
}

func NewNumberNode(numType string, num *Numeric) *NumberNode { return &NumberNode{num, numType} }

func (n *NumberNode) TestRepr() interface{} {
	var result []interface{}
	switch n.NumberType {
	case Integer:
		result = []interface{}{n.Type, n.Repr, float64(n.Integer), "integer"}
	case Float:
		result = []interface{}{n.Type, n.Repr, n.Float, "number"}
	default:
		result = []interface{}{"error", "bad-number"}
	}
	if n.Type == "dimension" {
		result = append(result, n.Unit)
	}
	return result
}

type TokenNode struct{ *Token }

func (n *TokenNode) str(nodeType string) interface{} {
	return []interface{}{nodeType, n.Value.(string)}
}

func (n *TokenNode) TestRepr() interface{} {
	switch n.TokenType {
	case WhitespaceToken:
		return " "
	case DelimToken:
		return string(n.Token.Value.(rune))
	case IdentToken:
		return []interface{}{"ident", string(n.Token.Value.(Identifier))}
	case AtKeywordToken:
		return n.str("at-keyword")
	case CDOToken:
		return "<!--"
	case CDCToken:
		return "-->"
	case StringToken:
		return n.str("string")
	case BadStringToken:
		return []interface{}{"error", "bad-string"}
	case UrlToken:
		return n.str("url")
	case BadUrlToken:
		return []interface{}{"error", "bad-url"}
	case UnicodeRangeToken:
		ur := n.Token.Value.(UnicodeRange)
		return []interface{}{"unicode-range", float64(ur.Start), float64(ur.End)}
	case IncludeMatchToken:
		return "~="
	case DashMatchToken:
		return "|="
	case PrefixMatchToken:
		return "^="
	case SuffixMatchToken:
		return "$="
	case SubstringMatchToken:
		return "*="
	case ColumnToken:
		return "||"
	case CommaToken:
		return ","
	case ColonToken:
		return ":"
	case SemicolonToken:
		return ";"
	default:
		return []interface{}{"error", n.TokenType.String()}
	}
}

func NewTokenNode(token *Token) Node {
	switch token.TokenType {
	case EOFToken:
		return NewEOFNode()
	case HashToken:
		switch id := token.Value.(type) {
		case string:
			return NewHashNode(id, true)
		case Identifier:
			return NewHashNode(string(id), false)
		}
	case NumberToken:
		return NewNumberNode("number", token.Value.(*Numeric))
	case DimensionToken:
		return NewNumberNode("dimension", token.Value.(*Numeric))
	case PercentageToken:
		return NewNumberNode("percentage", token.Value.(*Numeric))
	}
	return &TokenNode{token}
}

type DeclarationNode struct {
	Name      string
	Values    []Node
	Important bool
}

func NewDeclarationNode(name string, values []Node, important bool) *DeclarationNode {
	return &DeclarationNode{name, values, important}
}

func (n *DeclarationNode) TestRepr() interface{} {
	return []interface{}{"declaration", n.Name, nodeListTestRepr(n.Values), n.Important}
}

type Parser struct {
	tokenizer *Tokenizer
	current   *Token
	next      *Token
	reconsume bool
	debugOn   bool
}

func (p *Parser) debug(s ...interface{}) {
	if p.debugOn {
		fmt.Println(s...)
	}
}

func newParser(runeScanner io.RuneScanner, debugOn bool) *Parser {
	p := &Parser{tokenizer: NewTokenizer(runeScanner), debugOn: debugOn}
	p.current = p.tokenizer.ConsumeToken()
	p.debug("Consume:", p.current.String())
	p.next = p.tokenizer.ConsumeToken()
	return p
}

func NewParser(runeScanner io.RuneScanner) *Parser { return newParser(runeScanner, false) }

func NewDebugParser(runeScanner io.RuneScanner) *Parser { return newParser(runeScanner, true) }

func (p *Parser) consume1() {
	if p.reconsume {
		p.reconsume = false
		return
	}
	p.current = p.next
	p.next = p.tokenizer.ConsumeToken()
	p.debug("Consume:", p.current.String())
}

func (p *Parser) Consume1() *Token {
	p.consume1()
	return p.current
}

func (p *Parser) ParseListOfComponentValues() []Node {
	nodes := make([]Node, 0)
	for {
		next := p.consumeComponentValue()
		nodes = append(nodes, next)
		if nodeIsEOFOrError(next) {
			break
		}
		p.Consume1()
	}
	return nodes
}

func (p *Parser) consumeComponentValue() Node {
	tt := p.current.TokenType
	switch tt {
	case EOFToken:
		return NewEOFNode()
	case FunctionToken:
		return p.consumeFunction()
	case LCurlyToken:
		return p.consumeSimpleBlock(RCurlyToken)
	case LSquareToken:
		return p.consumeSimpleBlock(RSquareToken)
	case LParenToken:
		return p.consumeSimpleBlock(RParenToken)
	case RCurlyToken:
		return NewErrorNode(UnmatchedCurlyErr)
	case RSquareToken:
		return NewErrorNode(UnmatchedSquareErr)
	case RParenToken:
		return NewErrorNode(UnmatchedParenErr)
	default:
		return NewTokenNode(p.current)
	}
}

func (p *Parser) consumeSimpleBlock(delim TokenType) *BlockNode {
	values := make([]Node, 0)
	tt := p.Consume1().TokenType
	for tt != EOFToken && tt != ErrorToken && tt != delim {
		values = append(values, p.consumeComponentValue())
		tt = p.Consume1().TokenType
	}
	return NewBlockNode(delim, values...)
}

func (p *Parser) consumeFunction() Node {
	name := p.current.Value.(string)
	values := make([]Node, 0)
	tt := p.Consume1().TokenType
	for tt != EOFToken && tt != ErrorToken && tt != RParenToken {
		values = append(values, p.consumeComponentValue())
		tt = p.Consume1().TokenType
	}
	return NewFunctionNode(name, values...)
}

func (p *Parser) ParseDeclarationList() []Node {
	decls := make([]Node, 0)
	for p.current.TokenType != EOFToken {
		switch p.current.TokenType {
		case WhitespaceToken, SemicolonToken:
			p.Consume1()
		case AtKeywordToken:
			decls = append(decls, p.consumeAtRule())
		case IdentToken:
			decls = append(decls, p.consumeDeclaration())
		default:
			// FIXME: compliance with css3 tests, but not with standard
			// should just be consuming tokens, not component values
			for p.current.TokenType != EOFToken && p.current.TokenType != SemicolonToken {
				p.consumeComponentValue()
				p.Consume1()
			}
			p.Consume1()
			decls = append(decls, NewErrorNode(SyntaxErr))
		}
	}
	return decls
}

func (p *Parser) ParseDeclaration() Node {
	for p.current.TokenType == WhitespaceToken {
		p.Consume1()
	}
	if p.current.TokenType == EOFToken {
		return NewErrorNode(EmptyErr)
	}
	if p.current.TokenType != IdentToken {
		return NewErrorNode(SyntaxErr)
	}
	result := p.consumeDeclaration()
	if _, ok := result.(*DeclarationNode); ok && p.current.TokenType != EOFToken {
		return NewErrorNode(ExtraInputErr)
	}
	return result
}

func (p *Parser) consumeDeclaration() Node {
	name := string(p.current.Value.(Identifier))
	p.Consume1()
	for p.current.TokenType == WhitespaceToken {
		p.Consume1()
	}
	if p.current.TokenType != ColonToken {
		return NewErrorNode(SyntaxErr)
	}
	p.Consume1()
	values := make([]Node, 0)
	for p.current.TokenType != EOFToken && p.current.TokenType != SemicolonToken {
		// FIXME: compliance with css3 tests, but not with standard
		// should just be consuming tokens, not component values
		values = append(values, p.consumeComponentValue())
		p.Consume1()
	}
	if p.current.TokenType != EOFToken && p.current.TokenType != SemicolonToken {
		return NewErrorNode(ExtraInputErr)
	}
	p.Consume1()

	// Check if consumed values list ends with "!important"
	var important bool
	chopped, penult, last := nodeListLastTwoNonwhitespace(values)
	if nodeIsTokenType(penult, DelimToken) && penult.(*TokenNode).Value.(rune) == '!' {
		if nodeIsTokenType(last, IdentToken) {
			id := string(last.(*TokenNode).Value.(Identifier))
			if caseInsensitiveCompare(id, "important") {
				values = chopped
				important = true
			}
		}
	}

	// FIXME: compliance with css3 tests, but not with standard?
	for _, n := range values {
		if nodeIsTokenType(n, DelimToken) && n.(*TokenNode).Value.(rune) == '!' {
			return NewErrorNode(SyntaxErr)
		}
	}

	return NewDeclarationNode(name, values, important)
}

func (p *Parser) consumeAtRule() Node {
	name := p.current.Value.(string)
	prelude := make([]Node, 0)
	p.Consume1()
	for p.current.TokenType != EOFToken && p.current.TokenType != SemicolonToken {
		if p.current.TokenType == LCurlyToken {
			break
		}
		prelude = append(prelude, p.consumeComponentValue())
		p.Consume1()
	}
	var body []Node
	if p.current.TokenType == LCurlyToken {
		block := p.consumeSimpleBlock(RCurlyToken)
		if len(block.Values) > 0 {
			body = block.Values
		}
	}
	p.Consume1()
	return NewAtRuleNode(name, prelude, body)
}

func (p *Parser) ParseRule() Node {
	for p.current.TokenType == WhitespaceToken {
		p.Consume1()
	}
	if p.current.TokenType == EOFToken {
		return NewErrorNode(EmptyErr)
	}
	var result Node
	if p.current.TokenType == AtKeywordToken {
		result = p.consumeAtRule()
	} else {
		result = p.consumeQualifiedRule()
		// if nothing was returned, return a syntax error
	}
	p.Consume1()
	for p.current.TokenType == WhitespaceToken {
		p.Consume1()
	}
	if p.current.TokenType != EOFToken {
		return NewErrorNode(ExtraInputErr)
	}
	return result
}

func (p *Parser) ParseRuleList() []Node {
	return p.consumeRuleList(false)
}

func (p *Parser) consumeRuleList(toplevel bool) []Node {
	rules := make([]Node, 0)
	for p.current.TokenType != EOFToken {
		if p.current.TokenType == WhitespaceToken {
			p.Consume1()
			continue
		}
		if p.current.TokenType == CDOToken || p.current.TokenType == CDCToken {
			if !toplevel {
				rules = append(rules, p.consumeQualifiedRule())
			}
		} else if p.current.TokenType == AtKeywordToken {
			rules = append(rules, p.consumeAtRule())
		} else {
			rules = append(rules, p.consumeQualifiedRule())
		}
		p.Consume1()
	}
	return rules
}

func (p *Parser) consumeQualifiedRule() Node {
	var body []Node
	prelude := make([]Node, 0)
	for p.current.TokenType != LCurlyToken {
		if p.current.TokenType == EOFToken {
			return NewErrorNode(SyntaxErr)
		}
		prelude = append(prelude, p.consumeComponentValue())
		p.Consume1()
	}
	block := p.consumeSimpleBlock(RCurlyToken)
	if len(block.Values) > 0 {
		body = block.Values
	}
	return NewQualifiedRuleNode(prelude, body)
}

func (p *Parser) ParseStylesheet() []Node {
	return p.consumeRuleList(true)
}

func caseInsensitiveCompare(s1, s2 string) bool {
	r1 := []rune(s1)
	r2 := []rune(s2)
	if len(r1) != len(r2) {
		return false
	}
	for i, ch1 := range r1 {
		ch2 := r2[i]
		if ch1 >= 'A' && ch1 <= 'Z' {
			ch1 += 0x20
		}
		if ch2 >= 'A' && ch2 <= 'Z' {
			ch2 += 0x20
		}
		if ch1 != ch2 {
			return false
		}
	}
	return true
}

func toLower(s string) string {
	runes := make([]rune, 0, len(s))
	for _, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			ch += 0x20
		}
		runes = append(runes, ch)
	}
	return string(runes)
}
