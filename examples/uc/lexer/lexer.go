// generated by speak; DO NOT EDIT.

// Package lexer implements lexical analysis of the source language.
package lexer

import (
	"io"
	"io/ioutil"
	"regexp"

	"github.com/mewmew/speak/examples/uc/token"
	"github.com/pkg/errors"
)

// regstr specifies a regular expression for identifying the tokens of the input
// grammar.
const regstr = `^(('(?:\\n|a)')|([A-Z_a-z][0-9A-Z_a-z]*)|([0-9][0-9]*)|(!)|(!=)|(&&)|(\()|(\))|(\*)|(\+)|(,)|(-)|(/)|(;)|(<)|(<=)|(=)|(==)|(>)|(>=)|(\[)|(\])|(else)|(if)|(return)|(typedef)|(while)|(\{)|(\})|(//(?-s:.)*\n|#(?-s:.)*\n|/\*[^\*]*\*/)|([\t-\r ]))`

// reg is a compiled version of regstr with leftmost-longest matching enabled.
var reg *regexp.Regexp

func init() {
	// Compile regexp for identifying tokens and enforce leftmost-longest
	// matching.
	reg = regexp.MustCompile(regstr)
	reg.Longest()
}

// A Lexer lexes the source input into a slice of tokens.
type Lexer struct {
	// Source input.
	input []byte
	// Current position in the source input.
	pos int
}

// New returns a new scanner lexing from r.
func New(r io.Reader) (*Lexer, error) {
	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return NewFromBytes(input), nil
}

// Open returns a new scanner lexing from path.
func Open(path string) (*Lexer, error) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return NewFromBytes(input), nil
}

// NewFromString returns a new scanner lexing from input.
func NewFromString(input string) *Lexer {
	return NewFromBytes([]byte(input))
}

// NewFromBytes returns a new scanner lexing from input.
func NewFromBytes(input []byte) *Lexer {
	return &Lexer{input: input}
}

// Scan lexes and returns the next token of the source input.
func (l *Lexer) Scan() (*token.Token, error) {
	// Handle EOF.
	if l.pos >= len(l.input) {
		return nil, errors.WithStack(io.EOF)
	}
	input := l.input[l.pos:]
	// Identify token locations matching start of input.
	loc, err := tokenLocs(input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	n, id, err := locateTokens(input, loc)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	lit := input[:n]
	tok := &token.Token{
		Pos: l.pos,
		ID:  id,
		Lit: lit,
	}
	l.pos += n
	return tok, nil
}

// locateTokens searches for the longest token that match the start of the
// input.
func locateTokens(input []byte, loc []int) (n int, id token.ID, err error) {
	n = -1
	for i := 0; i < len(token.IDs); i++ {
		start := loc[2*i]
		if start == -1 {
			continue
		}
		if start != 0 {
			return 0, 0, errors.Errorf("invalid start index of token; expected 0, got %d", start)
		}
		end := loc[2*i+1]
		if n != -1 {
			return 0, 0, errors.Errorf("ambiguity detected; input matches both token %q and token %q", input[:n], input[:end])
		}
		n = end
		id = token.ID(i)
	}
	if n == -1 {
		// no matching token located.
		return 0, 0, errors.Errorf("unable to identify valid token at %q", input)
	}
	return n, id, nil
}

// tokenLocs returns start and end location of each token types that match the
// start of the input.
func tokenLocs(input []byte) ([]int, error) {
	loc := reg.FindSubmatchIndex(input)
	if loc == nil {
		// no submatch located.
		return nil, errors.Errorf("unable to identify valid token at %q", input)
	}
	// Validate submatch indices length; expecting two indices - start and end -
	// per submatch, and in total 2 + (number of tokens) submatches.
	got := len(loc)
	want := 2 * (2 + len(token.IDs))
	if got != want {
		return nil, errors.Errorf("invalid number of submatches; expected %d, got %d", want, got)
	}
	// Skip the first two submatches as they do not identify specific tokens.
	loc = loc[2*2:]
	return loc, nil
}
