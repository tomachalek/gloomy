// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
grammar:

R -> T
R -> TR
T -> F
T -> F?
T -> F*
T -> F|T ('|' is a consumed character here)
F -> atom
F -> (R)
F -> [L]

*/

package query

import (
	"fmt"
	"unicode"
)

func isAlnum(s rune) rune {
	if unicode.IsLetter(s) {
		return s
	}
	/*
		if s != '(' && s != ')' && s != '[' && s != ']' && s != '*' && s != '?' && s != '+' && s != '\u0003' {
			return s
		}*/
	return '\u0000'
}

type Parser struct {
	curr     int
	inputStr []rune
	stack    *altStack
	startAlt *atnstate
}

func (p *Parser) currChar() rune {
	if p.curr < len(p.inputStr) {
		return p.inputStr[p.curr]

	}
	return '\u0003'
}

func (p *Parser) fetchNextChar() {
	p.curr++
}

func (p *Parser) Parse(input string) error {
	p.curr = -1
	p.inputStr = []rune(input)
	p.startAlt = newState()
	p.stack.Push(p.startAlt)
	p.fetchNextChar()
	err := p.parseRegex()
	if err != nil {
		return err
	}
	if p.curr < len(p.inputStr) {
		return fmt.Errorf("Incomplete expression, position %d", p.curr)
	}
	return nil
}

func (p *Parser) GetAllPrefixes() []string {
	return p.startAlt.getAll()
}

// R -> TR'
func (p *Parser) parseRegex() error {
	var err error
	c := p.currChar()
	switch c {
	case isAlnum(c), '(', '[':
		err = p.parseTerm()
		if err != nil {
			break
		}
		err = p.parseRegexRest()
	default:
		return fmt.Errorf("Failed to parse [rule R]")
	}
	return err
}

// R' -> R
// R' -> eps
func (p *Parser) parseRegexRest() error {
	var err error
	c := p.currChar()
	switch c {
	case '(', '[', isAlnum(c):
		err = p.parseRegex()
		lastState := p.stack.Pop()
		p.stack.Peek().getLast().appendState(lastState)
	case ')':
		break
	case '\u0003':
		lastState := p.stack.Pop()
		p.stack.Peek().getLast().appendState(lastState)
		break // end of input
	default:
		err = fmt.Errorf("Failed to parse at %d [rule R']", p.curr)
	}
	return err
}

// T -> FT'
func (p *Parser) parseTerm() error {
	var err error
	c := p.currChar()
	switch c {
	case '(', '[', isAlnum(c):
		err = p.parseFactor()
		if err != nil {
			break
		}
		err = p.parseTermRest()
	default:
		err = fmt.Errorf("Failed to parse at %d [rule T]", p.curr)
	}
	return err
}

// T' -> eps
// T' -> F?
// T' -> F*
// T' -> F|T ('|' is a consumed character here)
func (p *Parser) parseTermRest() error {
	var err error
	c := p.currChar()
	switch c {
	case '*', '?':
		curr := p.stack.Pop()
		beg := newState()
		end := newState()
		beg.appendState(curr)
		curr.getLast().appendState(end)
		beg.appendState(end)
		p.stack.Push(beg)
		p.fetchNextChar()
	case '+':
		// we can't do anything reasonable here
		// in terms of prefix; we just stick
		// with the original term and no repeat.
		p.fetchNextChar()
	case '|':
		p.fetchNextChar()
		err = p.parseTerm()
		t2 := p.stack.Pop()
		t1 := p.stack.Pop()
		forkState := newState()
		forkState.appendState(t1)
		forkState.appendState(t2)
		joinState := newState()
		t1.getLast().appendState(joinState)
		t2.getLast().appendState(joinState)
		p.stack.Push(forkState)
	case '(', ')', '[', isAlnum(c), '\u0003':
		break
	default:
		err = fmt.Errorf("Parse error [nonterm T']")
	}
	return err
}

// F -> atom
// F -> (R)
// F -> [L]
func (p *Parser) parseFactor() error {
	var err error
	c := p.currChar()
	switch c {
	case '(':
		p.fetchNextChar()
		err = p.parseRegex()
		if err != nil {
			break
		}
		err = p.match(')')
	case '[':
		p.stack.Push(newState())
		p.fetchNextChar()
		err = p.parseList()
		if err != nil {
			break
		}
		err = p.match(']')
		if err == nil {
			joinState := newState()
			for _, c := range p.stack.Peek().children {
				c.appendState(joinState)
			}
		}
	case isAlnum(c):
		p.stack.Push(newRune(c))
		p.fetchNextChar()
	default:
		err = fmt.Errorf("Parse error [nonterm F]")
	}
	return err
}

func (p *Parser) parseList() error {
	var err error
	c := p.currChar()
	switch c {
	case isAlnum(c):
		p.stack.Peek().addRune(c)
		p.fetchNextChar()
		err = p.parseList()
	case ']':
		break
	default:
		err = fmt.Errorf("Parse error at position %d - incorrect character list", p.curr)
	}
	return err
}

func (p *Parser) match(c rune) error {
	if p.currChar() == c {
		p.fetchNextChar()
		return nil
	}
	return fmt.Errorf("Parse error - invalid input: %c, expected: %c", p.currChar(), c)
}

func NewParser() *Parser {
	return &Parser{stack: newAltStack()}
}
