package helpers

import (
	"regexp"
	"strings"
	"unicode"
)

type Spliter struct {
	Symbols         []*Symbol
	ExpressionsIndx string
	done            bool
	result          string
}

type Symbol struct {
	id             int
	value          string
	expressionType string
	isPriority     bool
	result         string
	op             string
}

var operations map[string]string = map[string]string{
	"+": "plus",
	"-": "minus",
	"*": "multiplication",
	"/": "division"}

func SplitExpression(expression string) ([]*Symbol, bool) {
	Symbols := []*Symbol{}
	lastNum := ""
	id := 1
	for q, v := range expression {
		if string(v) == "-" || string(v) == "+" {
			if string(v) == "-" {
				if q == 0 {
					lastNum += "(-"
					continue
				}
				if unicode.IsNumber([]rune(expression)[q+1]) && !unicode.IsNumber([]rune(expression)[q-1]) && string(expression[q-1]) != ")" {
					lastNum += "(-"
					continue
				}
			}
			if len(lastNum) != 0 {
				Symbols = append(Symbols, &Symbol{value: lastNum, expressionType: "num", isPriority: false, id: id})
				lastNum = ""
				id++
			}
			if string(v) == "+" {
				Symbols = append(Symbols, &Symbol{value: string(v), expressionType: "operation", isPriority: false, id: id})
				id++
			} else {
				Symbols = append(Symbols, &Symbol{value: string(v), expressionType: "operation", isPriority: false, id: id})
				id++
				continue
			}
			continue
		} else {
			if string(v) == "*" || string(v) == "/" {
				if len(lastNum) != 0 {
					Symbols = append(Symbols, &Symbol{value: lastNum, expressionType: "num", isPriority: false, id: id})
					lastNum = ""
					id++
				}
				Symbols = append(Symbols, &Symbol{value: string(v), expressionType: "operation", isPriority: true, id: id})
				id++
				continue
			}
		}
		if string(v) != "(" {
			lastNum += string(v)
		}
	}
	if len(lastNum) != 0 {
		Symbols = append(Symbols, &Symbol{value: lastNum, expressionType: "num", isPriority: false, id: id})
		lastNum = ""
	}
	for _, v := range Symbols {
		if string(v.value[0]) == "(" && string(v.value[len(v.value)-1]) != ")" {
			v.value += ")"
		}
	}
	if len(Symbols) == 1 {
		return Symbols, false
	}
	return Symbols, true
}

func NewSpliter(expression string) *Spliter {
	spliter := &Spliter{}
	symbols, ok := SplitExpression(expression)
	if !ok {
		spliter.done = true
		spliter.result = symbols[0].value
	}
	spliter.Symbols = symbols
	return spliter
}

func (s *Spliter) Update(expression string) {
	symbols, ok := SplitExpression(expression)
	if !ok {
		s.done = true
		s.result = symbols[0].value
	}
	s.Symbols = symbols
}

func (s *Spliter) getNextOperation(id int) *Symbol {
	if id+1 >= len(s.Symbols) {
		return &Symbol{}
	}
	for _, v := range s.Symbols[id+1:] {
		if v.expressionType == "operation" {
			return v
		}
	}
	return &Symbol{}
}

func (s *Spliter) getLastOperation(id int) *Symbol {
	for i := id - 2; i >= 0; i-- {
		if s.Symbols[i].expressionType == "operation" {
			return s.Symbols[i]
		}
	}
	return &Symbol{}
}

func (s *Spliter) Split() {
	id := 1
	new := []*Symbol{}
	for i := 0; i < len(s.Symbols); i++ {
		if s.Symbols[i].isPriority && !s.getLastOperation(s.Symbols[i].id).isPriority {
			new = append(new[:id-1], &Symbol{id: id, expressionType: "calculation", value: s.Symbols[i-1].value + s.Symbols[i].value + s.Symbols[i+1].value, op: operations[s.Symbols[i].value]})
			id++
			i += 2
			if i+1 < len(s.Symbols) {
				new = append(new, &Symbol{id: id, expressionType: s.Symbols[i].expressionType, value: s.Symbols[i].value})
				id++
			}
			continue
		}
		if s.Symbols[i].expressionType == "operation" && !s.getLastOperation(s.Symbols[i].id).isPriority && !s.getNextOperation(s.Symbols[i].id).isPriority {
			if s.getLastOperation(s.Symbols[i].id).value == "-" {
				if s.Symbols[i].value == "-" {
					new = append(new[:id-1], &Symbol{id: id, expressionType: "calculation", value: s.Symbols[i-1].value + "+" + s.Symbols[i+1].value, op: "plus"})
				} else {
					new = append(new[:id-1], &Symbol{id: id, expressionType: "calculation", value: s.Symbols[i-1].value + "-" + s.Symbols[i+1].value, op: "minus"})
				}
			} else {
				new = append(new[:id-1], &Symbol{id: id, expressionType: "calculation", value: s.Symbols[i-1].value + s.Symbols[i].value + s.Symbols[i+1].value, op: operations[s.Symbols[i].value]})
			}
			i += 2
			if i+1 < len(s.Symbols) {
				new = append(new, &Symbol{id: id, expressionType: s.Symbols[i].expressionType, value: s.Symbols[i].value})
				id++
			}
			id++
			continue
		}
		if (s.Symbols[i].expressionType == "num" && (!s.getNextOperation(s.Symbols[i].id-1).isPriority && s.getNextOperation(s.Symbols[i].id).isPriority || s.getLastOperation(s.Symbols[i].id).isPriority)) || i+1 == len(s.Symbols) {
			new = append(new, &Symbol{id: id, expressionType: s.Symbols[i].expressionType, value: s.Symbols[i].value})
			id++
		}
		if s.Symbols[i].expressionType == "operation" {
			new = append(new, &Symbol{id: id, expressionType: s.Symbols[i].expressionType, value: s.Symbols[i].value})
			id++
		}
	}
	s.Symbols = new
}

func IsValidExpression(expression string) bool {
	expression = strings.ReplaceAll(expression, " ", "")
	if matched, _ := regexp.MatchString(`^[*/+]|[*\/+]$`, expression); matched {
		return false
	}
	re := regexp.MustCompile(`\/\s*-?0+(\s*[\+\-\*\/]|$)`)
	if re.MatchString(expression) {
		return false
	}
	pattern := `^-?\d+(\.\d+)?([+\-*/]-?\d+(\.\d+)?)*$`
	return regexp.MustCompile(pattern).MatchString(expression)
}
