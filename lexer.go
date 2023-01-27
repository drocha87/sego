package main

import (
	"errors"
	"strings"
	"unicode"
)

type Lexer struct {
	content []rune
}

func NewLexer(s string) *Lexer {
	return &Lexer{content: []rune(s)}
}

func (l *Lexer) TrimLeft() {
	n := 0
	for n < len(l.content) && unicode.IsSpace(l.content[n]) {
		n++
	}
	l.content = l.content[n:]
}

func (l *Lexer) Chop(n int) []rune {
	result := make([]rune, n)
	for i, r := range l.content {
		result[i] = r
		if i == n-1 {
			break
		}
	}
	l.content = l.content[n:]
	return result
}

func (l *Lexer) ChopWhile(predicate func(rune) bool) []rune {
	n := 0
	for n < len(l.content) && predicate(l.content[n]) {
		n += 1
	}
	return l.Chop(n)
}

func (l *Lexer) IsEmpty() bool {
	return len(l.content) <= 0
}

func (l *Lexer) NextToken() (string, error) {
	l.TrimLeft()

	if l.IsEmpty() {
		return "", errors.New("EOF")
	}
	if unicode.IsNumber(l.content[0]) {
		return string(l.ChopWhile(unicode.IsNumber)), nil
	}
	if unicode.IsLetter(l.content[0]) {
		s := string(l.ChopWhile(func(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }))
		return strings.ToUpper(s), nil
	}
	return string(l.Chop(1)), nil
}
