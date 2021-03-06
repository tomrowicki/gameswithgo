package apt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type tokenType int

const (
	openParen tokenType = iota
	closeParen
	op
	constant
)

type token struct {
	typ   tokenType
	value string
}

type lexer struct {
	input  string
	start  int
	pos    int
	width  int
	tokens chan token
}

func stringToNode(s string) Node {
	switch s {
	case "+":
		return NewOpPlus()
	case "-":
		return NewOpMinus()
	case "*":
		return NewOpMult()
	case "/":
		return NewOpDiv()
	case "Atan2":
		return NewOpAtan2()
	case "Atan":
		return NewOpAtan()
	case "Cos":
		return NewOpCos()
	case "Sin":
		return NewOpSin()
	case "SimplexNoise":
		return NewOpNoise()
	case "Lerp":
		return NewOpLerp()
	case "X":
		return NewOpX()
	case "Y":
		return NewOpY()
	case "Picture":
		return NewOpPicture()
	default:
		panic("Can't determine node from token: " + s)
	}
}

func parse(tokens chan token, parent Node) Node {
	for {
		token, ok := <-tokens
		if !ok {
			fmt.Println()
			panic("no more tokens")
		}
		switch token.typ {
		case op:
			n := stringToNode(token.value)
			n.SetParent(parent)
			for i := range n.GetChildren() {
				n.GetChildren()[i] = parse(tokens, n)
			}
			return n
		case constant:
			n := NewOpConstant()
			n.SetParent(parent)
			v, err := strconv.ParseFloat(token.value, 32)
			if err != nil {
				panic("Error while parsing constant op.")
			}
			n.value = float32(v)
			return n
		case closeParen, openParen:
			continue
		}
	}
	return nil
}

const eof rune = -1

type stateFunc func(*lexer) stateFunc

func BeginLexing(s string) Node {
	l := &lexer{input: s, tokens: make(chan token, 100)}
	go l.run()
	return parse(l.tokens, nil)
}

func (l *lexer) run() {
	for state := determineToken(l); state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func determineToken(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case isWhiteSpace(r):
			l.ignore()
		case r == '(':
			l.emit(openParen)
		case r == ')':
			l.emit(closeParen)
		case isStartOfNumber(r):
			return lexNumber
		case r == eof || r == rune(65533): // some unknown (linux-related?) rune messes things up and initiates an endless loop
			return nil
		default:
			// only operators remain
			return lexOp
		}
	}
}

func lexOp(l *lexer) stateFunc {
	l.acceptRun("+-/*abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	l.emit(op)
	return determineToken
}

func lexNumber(l *lexer) stateFunc {
	l.accept("-.") // beginning of a number
	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits) // parsing decimal places
	}

	if l.input[l.start:l.pos] == "-" {
		l.emit(op)
	} else {
		l.emit(constant)
	}

	return determineToken
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func isWhiteSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r'
}

// start as in the beginning of a floating point number
func isStartOfNumber(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-' || r == '.'
}

func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos > len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) peek() (r rune) {
	r, _ = utf8.DecodeLastRuneInString(l.input[l.pos:])
	return r
}
