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
)

func isAlnum(s rune) rune {
	if s != '(' && s != ')' && s != '[' && s != ']' && s != '*' && s != '?' && s != '+' && s != '\u0003' {
		return s
	}
	return '\u0000'
}

type Parser struct {
	curr     int
	inputStr []rune
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
	p.fetchNextChar()
	return p.parseRegex()
}

func (p *Parser) parseRegex() error {
	c := p.currChar()
	fmt.Printf("parse regex [%c] -->\n", c)
	switch c {
	case isAlnum(c), '(', '[':
		p.parseTerm()
		p.parseRegexRest()
	default:
		return fmt.Errorf("Failed to parse")
	}
	return nil
}

func (p *Parser) parseRegexRest() error {
	c := p.currChar()
	fmt.Printf("parse regex rest [%c] -->\n", c)
	switch c {
	case '(', '[', isAlnum(c):
		p.parseRegex()
	case ')':
		break
	case '\u0003':
		break // end of input
	default:
		return fmt.Errorf("Failed to parse")
	}
	return nil
}

func (p *Parser) parseTerm() error {
	c := p.currChar()
	fmt.Printf("parseTerm [%c] ->\n", c)
	switch c {
	case '(', '[', isAlnum(c):
		p.parseFactor()
		p.parseTermRest()
	default:
		return fmt.Errorf("Parse error... TODO")
	}
	return nil
}

func (p *Parser) parseTermRest() error {
	c := p.currChar()
	fmt.Printf("parse term rest [%c] ->\n", c)
	switch c {
	case '*', '+', '?':
		fmt.Print("-- wildcard ---")
		p.fetchNextChar()
	case '|':
		p.parseTerm()
	case '(', ')', isAlnum(c), '\u0003':
		break
	default:
		return fmt.Errorf("Parse error ... TODO")
	}
	return nil
}

func (p *Parser) parseFactor() {
	fmt.Print("parseFactor ->\n")
	c := p.currChar()
	switch c {
	case '(':
		p.fetchNextChar()
		p.parseRegex()
		p.match(')')
	case '[':
		p.fetchNextChar()
		p.parseList()
		p.match(']')
	case isAlnum(c):
		fmt.Printf("Terminal: %c\n\n", c)
		p.fetchNextChar()
	}
}

func (p *Parser) parseList() error {
	c := p.currChar()
	switch c {
	case isAlnum(c):
		fmt.Printf("list item [%c]\n", c)
		p.fetchNextChar()
		p.parseList()
	case ']':
		break
	default:
		return fmt.Errorf("Parse error")
	}
	return nil
}

func (p *Parser) match(c rune) error {
	if p.currChar() == c {
		fmt.Printf("MATCH %c (input: %c)\n", p.currChar(), c)
		p.fetchNextChar()
		return nil
	}
	return fmt.Errorf("Invalid input: %c, expected: %c", p.currChar(), c)
}

func NewParser() *Parser {
	return &Parser{}
}
