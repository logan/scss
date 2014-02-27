package css3

import (
	"bytes"
	"errors"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	readRuneFailer   rune = '!'
	unreadRuneFailer rune = '?'
)

var testRuneScannerError = errors.New("forced")

type testRuneScanner struct {
	io.RuneScanner
	last rune
}

func (t *testRuneScanner) ReadRune() (rune, int, error) {
	r, s, e := t.RuneScanner.ReadRune()
	t.last = r
	if r == readRuneFailer {
		e = testRuneScannerError
	}
	return r, s, e
}

func (t *testRuneScanner) UnreadRune() error {
	if t.last == unreadRuneFailer {
		return testRuneScannerError
	}
	return t.RuneScanner.UnreadRune()
}

func TestScanner(t *testing.T) {
	scan := func(input string) *Scanner {
		return NewScanner(bytes.NewReader([]byte(input)))
	}

	errscan := func(input string) *Scanner {
		s := NewScanner(&testRuneScanner{RuneScanner: bytes.NewReader([]byte(input))})
		s.Consume(1)
		return s
	}

	shouldHaveState := func(actual interface{}, expected ...interface{}) string {
		s := actual.(*Scanner)
		var expErr error
		if expected[0] == nil {
			expErr = nil
		} else {
			expErr = expected[0].(error)
		}
		expCur := expected[1].(rune)
		expNext := expected[2].(rune)
		expPeek := expected[3].(string)
		if msg := ShouldEqual(s.Error(), expErr); msg != "" {
			return msg
		}
		if msg := ShouldEqual(s.Current(), expCur); msg != "" {
			return msg
		}
		if msg := ShouldEqual(s.Next(), expNext); msg != "" {
			return msg
		}
		if msg := ShouldEqual(s.PeekString(), expPeek); msg != "" {
			return msg
		}
		return ""
	}

	Convey("preprocessing", t, func() {
		s := scan("a b\rc\r\nd\fe\000f\r")
		runes := make([]rune, 0)
		for {
			s.Consume(1)
			if s.Current() < 0 {
				break
			}
			runes = append(runes, s.Current())
		}
		So(string(runes), ShouldEqual, "a b\nc\nd\ne\ufffdf\n")
	})

	Convey("lookahead and reconsume", t, func() {
		s := scan("abc\r\ndef")

		s.Consume(1)
		So(s.Peek3(), ShouldResemble, []rune{'b', 'c', '\n'})
		So(s, shouldHaveState, nil, 'a', 'b', "bc\n")

		s.Consume(1)
		So(s, shouldHaveState, nil, 'b', 'c', "c\nd")

		s.Consume(3)
		So(s, shouldHaveState, nil, 'd', 'e', "ef")

		s.Reconsume()
		s.Consume(1)
		So(s, shouldHaveState, nil, 'd', 'e', "ef")

		s.Reconsume()
		s.Consume(2)
		So(s, shouldHaveState, nil, 'e', 'f', "f")

		s.Consume(1)
		So(s, shouldHaveState, nil, 'f', EOFRune, "")

		s.Consume(1)
		So(s, shouldHaveState, nil, EOFRune, EOFRune, "")
	})

	Convey("error handling", t, func() {
		// fill boundary conditions
		So(errscan(""), shouldHaveState, nil, EOFRune, EOFRune, "")
		So(errscan("a"), shouldHaveState, nil, 'a', EOFRune, "")
		So(errscan("ab"), shouldHaveState, nil, 'a', 'b', "b")
		So(errscan("abc"), shouldHaveState, nil, 'a', 'b', "bc")
		So(errscan("abcd"), shouldHaveState, nil, 'a', 'b', "bcd")
		So(errscan("!abcd"), shouldHaveState, testRuneScannerError, ErrorRune, ErrorRune, "")
		So(errscan("a!bcd"), shouldHaveState, testRuneScannerError, 'a', ErrorRune, "")
		So(errscan("ab!cd"), shouldHaveState, testRuneScannerError, 'a', 'b', "b")
		So(errscan("abc!d"), shouldHaveState, testRuneScannerError, 'a', 'b', "bc")
		So(errscan("abcd!"), shouldHaveState, nil, 'a', 'b', "bcd")

		// interrupted \r\n
		So(errscan("\r"), shouldHaveState, nil, '\n', EOFRune, "")
		So(errscan("\r!\n"), shouldHaveState, testRuneScannerError, '\n', ErrorRune, "")
		So(errscan("\r?\n"), shouldHaveState, testRuneScannerError, '\n', ErrorRune, "")
	})
}
