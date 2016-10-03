package version

import (
	"bufio"
	"bytes"
	"io"
)

// Scanner represents a lexical scanner for ranges.
type scanner struct {
	r *reader
}

// NewScanner returns a new instance of Scanner.
func newRangeScanner(r io.Reader) *scanner {
	return &scanner{r: &reader{r: bufio.NewReader(r)}}
}

// Scan returns the next token and position from the underlying reader.
func (s *scanner) Scan() (tok Token, pos Pos, lit string) {
	ch0, pos := s.r.read()

	// Recognize and consume complex segments
	if isWhitespace(ch0) {
		return s.scanWhitespace()
	} else if isDigit(ch0) {
		return s.scanVersion()
	} else if ch0 == eof {
		return EOF, pos, ""
	}

	// Try to match two char operators
	ch1, _ := s.r.read()
	switch string([]rune{ch0, ch1}) {
	case "~>":
		return TILDE, pos, ""
	case "||":
		return OR, pos, ""
	case "==":
		return EQ, pos, ""
	case ">=":
		return GTE, pos, ""
	case "<=":
		return LTE, pos, ""
	case "!=":
		fallthrough
	case "<>":
		return NEQ, pos, ""
	}

	// Rewind and try just one.
	s.r.unread()
	switch ch0 {
	case '~':
		return TILDE, pos, ""
	case '^':
		return CARET, pos, ""
	case '!':
		return NEQ, pos, ""
	case '=':
		return EQ, pos, ""
	case '>':
		return GT, pos, ""
	case '<':
		return LT, pos, ""
	}

	return INVALID, pos, string(ch0)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *scanner) scanWhitespace() (tok Token, pos Pos, lit string) {
	pos, lit = s.scanWhile(isWhitespace)
	return WS, pos, lit
}

// scanNumber consumes anything version-like that starts with a number
func (s *scanner) scanVersion() (tok Token, pos Pos, lit string) {
	pos, lit = s.scanWhile(isVersion)
	return VERSION, pos, lit
}

// Generic helper for scanning while isMatch(ch) returns true
func (s *scanner) scanWhile(isMatch func(rune) bool) (pos Pos, lit string) {
	var buf bytes.Buffer
	ch, pos := s.r.curr()
	for ; isMatch(ch); ch = s.r.readRune() {
		_, _ = buf.WriteRune(ch)
	}
	if ch != eof {
		s.r.unread()
	}
	return pos, buf.String()
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == ','
}

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// isVersion includes digits, letters, hyphens, etc
func isVersion(ch rune) bool {
	return isDigit(ch) || isLetter(ch) || ch == '.' || ch == '-' || ch == '+'
}

// reader represents a buffered rune reader used by the scanner.
// It provides a fixed-length circular buffer that can be unread.
type reader struct {
	r        io.RuneScanner
	eof      bool // true if we read eof
	pos      Pos  // last read position
	current  rune // last read rune
	previous rune // and one before
}

// curr returns the last read character and position.
func (r *reader) curr() (ch rune, pos Pos) {
	return r.current, r.pos - 1
}

func (r *reader) readRune() rune {
	ch, _ := r.read()
	return ch
}

// read reads the next rune from the reader.
func (r *reader) read() (ch rune, pos Pos) {
	ch, _, err := r.r.ReadRune()
	if err != nil {
		ch = eof
	}

	if !r.eof {
		r.eof = (ch == eof)
		r.pos++
	}

	// Save current and previous runes
	r.previous, r.current = r.current, ch
	return r.curr()
}

// Unread pushes the previously read rune back onto the buffer.
func (r *reader) unread() {
	if r.previous == eof {
		panic("Double unread")
	}
	if r.r.UnreadRune() == nil {
		r.current, r.previous = r.previous, eof
		r.pos--
	}
}

// eof is a marker code point to signify that the reader can't read any more.
const eof = rune(0)

type Pos int

//======== Token types ========

type Token int

const (
	// INVALID Token, EOF, WS are Special InfluxQL tokens.
	INVALID Token = iota
	EOF
	WS

	VERSION // 1.2.3-bla
	TILDE   // ~
	CARET   // ^

	OR  // OR
	EQ  // ==
	NEQ // !=
	LT  // <
	LTE // <=
	GT  // >
	GTE // >=
)

var tokens = [...]string{
	INVALID: "INVALID",
	EOF:     "EOF",
	WS:      "WS",

	VERSION: "VERSION",
	TILDE:   "~",
	CARET:   "^",

	OR:  "OR",
	EQ:  "=",
	NEQ: "!=",
	LT:  "<",
	LTE: "<=",
	GT:  ">",
	GTE: ">=",
}

// String returns the string representation of the token.
func (tok Token) String() string {
	if tok < 0 || tok >= Token(len(tokens)) {
		return ""
	}
	return tokens[tok]
}
