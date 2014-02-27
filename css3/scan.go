package css3

import (
	"io"
)

const EOFRune rune = -1
const ErrorRune rune = -2

type codePoint struct {
	rune
	error
}

type Scanner struct {
	error
	preprocessor

	current   rune
	next      []rune
	reconsume bool
}

func NewScanner(runeScanner io.RuneScanner) *Scanner {
	return &Scanner{preprocessor: preprocessor{RuneScanner: runeScanner}}
}

func (s *Scanner) nextRune() (rune, error) {
	return s.preprocessor.nextRune()
}

func (s *Scanner) fill() {
	if s.next == nil {
		s.current, s.error = s.nextRune()
		s.next = make([]rune, 0, 3)
	}
	for len(s.next) < 3 {
		ch, err := s.nextRune()
		s.next = append(s.next, ch)
		s.error = err
	}
}

func (s *Scanner) consume1() {
	if s.reconsume {
		s.reconsume = false
		return
	}
	if s.next == nil {
		s.fill()
	} else {
		s.current = s.next[0]
		s.next[0] = s.next[1]
		s.next[1] = s.next[2]
		s.next[2], s.error = s.nextRune()
	}
}

func (s *Scanner) Error() error  { return s.error }
func (s *Scanner) Reconsume()    { s.reconsume = true }
func (s *Scanner) Current() rune { return s.current }
func (s *Scanner) Next() rune    { return s.next[0] }
func (s *Scanner) Peek3() []rune { return s.next }

func (s *Scanner) Consume1() rune {
	s.consume1()
	return s.current
}

func (s *Scanner) Consume(n int) {
	for i := 0; i < n; i++ {
		s.consume1()
	}
}

func (s *Scanner) PeekString() string {
	for i := 0; i < len(s.next); i++ {
		if s.next[i] < 0 {
			return string(s.next[:i])
		}
	}
	return string(s.next)
}

type preprocessor struct {
	error
	io.RuneScanner
	eof bool
}

func (p *preprocessor) nextRune() (rune, error) {
	if p.error != nil {
		return ErrorRune, p.error
	}
	if p.eof {
		return EOFRune, nil
	}

	next, _, err := p.RuneScanner.ReadRune()
	if err != nil {
		if err == io.EOF {
			p.eof = true
			return EOFRune, nil
		}
		p.error = err
		return ErrorRune, err
	}
	if next == '\r' {
		next, _, err = p.RuneScanner.ReadRune()
		if err != nil {
			if err == io.EOF {
				p.eof = true
			} else {
				p.error = err
			}
			return '\n', nil
		} else if next != '\n' {
			p.error = p.RuneScanner.UnreadRune()
			return '\n', nil
		}
		next = '\n'
	} else if next == '\f' {
		next = '\n'
	} else if next == 0 {
		next = '\ufffd'
	}
	return next, nil
}
