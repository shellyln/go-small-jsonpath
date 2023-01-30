package jsonpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type astType int

const (
	astType_NameIndexer astType = iota + 1
	astType_NumberIndexer
	astType_Function
)

type ast struct {
	typ   astType
	name  string
	index int
}

type JSONValueType int

const (
	Type_Invalid JSONValueType = iota
	Type_Null
	Type_Number
	Type_String
	Type_Object
	Type_Array
)

type parsedJSON struct {
	typ   JSONValueType
	value interface{}
}

type CompiledJSONPath struct {
	asts []ast
}

func newParsedJSON() *parsedJSON {
	return &parsedJSON{}
}

func ReadString(src string) (*parsedJSON, error) {
	p := newParsedJSON()
	var err error

	src2 := strings.TrimSpace(src)
	if src == "" {
		return nil, errors.New("ReadString: Source is empty")
	}

	p.value = nil

	// NOTE: NaN and Infinity are not valid JSON
	switch src[0] {
	case 'n':
		if src != "null" {
			return nil, fmt.Errorf("ReadString: Unrecognised tokens appeared: Pos=%v, %v", 0, string(src[0:]))
		}
		p.typ = Type_Null
	case '{':
		dst := make(map[string]interface{})
		err = json.Unmarshal([]byte(src2), &dst)
		p.typ = Type_Object
		p.value = dst
	case '[':
		dst := make([]interface{}, 0, 100)
		err = json.Unmarshal([]byte(src2), &dst)
		p.typ = Type_Array
		p.value = dst
	case '"':
		dst := ""
		err = json.Unmarshal([]byte(src2), &dst)
		p.typ = Type_String
		p.value = dst
	default:
		dst := float64(0.0)
		err = json.Unmarshal([]byte(src2), &dst)
		p.typ = Type_Number
		p.value = dst
	}

	if err != nil {
		return nil, err
	} else {
		return p, nil
	}
}

func (p parsedJSON) Root() interface{} {
	return p.value
}

func Compile(path string) (*CompiledJSONPath, error) {
	return compileCore([]rune(path), '$')
}

func compileCore(src []rune, root rune) (*CompiledJSONPath, error) {
	if src[0] != root {
		return nil, fmt.Errorf("compileCore: Path should be starts with '%v': Pos=%v, %v", string(root), 0, string(src[0]))
	}

	length := len(src)
	asts := make([]ast, 0, 20)
	var start, end int
	var err error
	var name string

	for i := 1; i < length; i++ {
		ch := src[i]

		if unicode.IsSpace(ch) || unicode.IsControl(ch) {
			end, err = skipSpaces(src, i+1)
			if err != nil {
				return nil, fmt.Errorf("compileCore: Unexpected termination: Pos=%v, %v", i, src[i:])
			}
			i = end - 1

		} else {
			switch ch {
			case '[':
				// number indexer / name indexer
				end, err = skipSpaces(src, i+1)
				if err != nil || end == length {
					return nil, fmt.Errorf("compileCore: Unexpected termination in the '[' bracket: Pos=%v", i)
				}
				start = end

				// TODO: %#name : number indexer variable
				// TODO: %name  : name indexer variable

				if '0' <= src[start] && src[start] <= '9' || src[start] == '-' {
					end, err = parseNumber(src, start)
					if err != nil {
						return nil, fmt.Errorf("compileCore: Bad number expression: Pos=%v, %v", start, src[start:])
					}
					num, err := strconv.ParseInt(string(src[start:end]), 10, 64)
					if err != nil {
						return nil, fmt.Errorf("compileCore: Integer cannot be parsed: Pos=%v, %v", start, src[start:end])
					}
					asts = append(asts, ast{
						typ:   astType_NumberIndexer,
						index: int(num),
					})
				} else {
					ch2 := src[start]

					switch ch2 {
					case '\'', '"':
						// quoted name
						name, end, err = parseQuotedName(src, ch2, start+1)
						if err != nil {
							return nil, fmt.Errorf("compileCore: Bad quoted name expression: Pos=%v, %v", start, src[start:])
						}
						asts = append(asts, ast{
							typ:  astType_NameIndexer,
							name: name,
						})
					default:
						return nil, fmt.Errorf("compileCore: Bad quoted name expression: Pos=%v, %v", start, src[start:])
					}
				}

				end, err = skipSpaces(src, end)
				if err != nil || end == length {
					return nil, fmt.Errorf("compileCore: Unexpected termination in the '[' bracket: Pos=%v", start)
				}

				if src[end] != ']' {
					return nil, fmt.Errorf("compileCore: '[' bracket is not closed: Pos=%v, %v", end, src[end:])
				}
				i = end // end is ']'

			case '.':
				end, err = skipSpaces(src, i+1)
				if err != nil || end == length {
					return nil, fmt.Errorf("compileCore: Unexpected termination after '.': Pos=%v", i)
				}
				start = end
				ch2 := src[start]

				switch ch2 {
				case '(':
					// function
					end, err = skipSpaces(src, start+1)
					if err != nil {
						return nil, fmt.Errorf("compileCore: Unexpected termination in the '(' parenthesis: Pos=%v", i)
					}
					start = end

					name, end, err = parseBareName(src, start)
					if err != nil {
						return nil, fmt.Errorf("compileCore: Bad function name expression: Pos=%v, %v", start, src[start:])
					}
					asts = append(asts, ast{
						typ:  astType_Function,
						name: string(name),
					})

					end, err = skipSpaces(src, end)
					if err != nil || end == length {
						return nil, fmt.Errorf("compileCore: Unexpected termination in the '(' parenthesis: Pos=%v", start)
					}

					if src[end] != ')' {
						return nil, fmt.Errorf("compileCore: '(' parenthesis is not closed: Pos=%v, %v", end, src[end:])
					}
					i = end // end is ')'

				default:
					// bare name
					name, end, err = parseBareName(src, start)
					if err != nil {
						return nil, fmt.Errorf("compileCore: Bad name expression: Pos=%v, %v", start, src[start:])
					}
					asts = append(asts, ast{
						typ:  astType_NameIndexer,
						name: name,
					})

					end, err = skipSpaces(src, end)
					if err != nil {
						return nil, fmt.Errorf("compileCore: Bad name expression: Pos=%v, %v", start, src[start:])
					}
					i = end - 1
				}

			default:
				return nil, fmt.Errorf("compileCore: Unexpected character appeared: Pos=%v, %v", i, src[i:])
			}
		}
	}

	return &CompiledJSONPath{
		asts: asts,
	}, nil
}

func (p *CompiledJSONPath) Query(pjson *parsedJSON) (interface{}, error) {
	if pjson.typ == Type_Invalid {
		return nil, errors.New("Query: JSON is not read")
	}

	v := pjson.value
	var ok bool

	for i, a := range p.asts {
		if v == nil {
			return nil, fmt.Errorf("Query: Nil referenced: Level=%v", i)
		}

		switch z := v.(type) {
		case map[string]interface{}:
			switch a.typ {
			case astType_NameIndexer:
				v, ok = z[a.name]
				if !ok {
					return nil, fmt.Errorf("Query: Property %v does not exist in the object: Level=%v", a.name, i)
				}
			case astType_NumberIndexer:
				return nil, fmt.Errorf("Query: Object cannot be accessed by number: Level=%v, %v", i, a.index)
			case astType_Function:
				return nil, fmt.Errorf("Query: Object cannot be accessed by function: Level=%v, %v", i, a.name)
			}

		case []interface{}:
			length := len(z)
			switch a.typ {
			case astType_NameIndexer:
				return nil, fmt.Errorf("Query: Array cannot be accessed by name: Level=%v, %v", i, a.name)
			case astType_NumberIndexer:
				idx := a.index
				if idx < 0 {
					idx = length - idx
				}
				if length <= idx {
					return nil, fmt.Errorf("Query: Index out of range: Level=%v, length=%v, %v", i, length, a.index)
				}
				v = z[idx]
			case astType_Function:
				switch a.name {
				case "length":
					v = length
				case "first":
					if length == 0 {
						return nil, fmt.Errorf("Query: Index out of range: Level=%v, length=%v, (first)", i, length)
					}
					v = z[0]
				case "last":
					if length == 0 {
						return nil, fmt.Errorf("Query: Index out of range: Level=%v, length=%v, (last)", i, length)
					}
					v = z[length-1]
				default:
					return nil, fmt.Errorf("Query: Undefined function name: Level=%v, %v", i, a.name)
				}
			}

		default:
			return nil, fmt.Errorf("Query: Unexpected data type appeared: Level=%v", i)
		}
	}

	return v, nil
}

func (p *CompiledJSONPath) QueryAsStringOrZero(pjson *parsedJSON) string {
	v, err := p.Query(pjson)
	if err != nil {
		return ""
	}

	ret, ok := v.(string)
	if !ok {
		return ""
	}
	return ret
}

func (p *CompiledJSONPath) QueryAsNumberOrZero(pjson *parsedJSON) float64 {
	v, err := p.Query(pjson)
	if err != nil {
		return 0
	}

	ret, ok := v.(float64)
	if !ok {
		return 0
	}
	return ret
}

func skipSpaces(src []rune, start int) (int, error) {
	length := len(src)

	for i := start; i < length; i++ {
		ch := src[i]
		if !unicode.IsSpace(ch) && !unicode.IsControl(ch) {
			return i, nil
		}
	}
	return length, nil
}

func parseQuotedName(src []rune, cc rune, start int) (string, int, error) {
	length := len(src)
	buf := make([]rune, 0, 32)

	for i := start; i < length; i++ {
		ch := src[i]
		switch ch {
		case cc:
			return string(buf), i + 1, nil

		case '\\':
			if i+1 == length {
				return "", start, fmt.Errorf("parseQuotedName: Unexpected termination: Pos=%v", i)
			}

			switch src[i+1] {
			case '\\', '"', '\'', '`':
				buf = append(buf, src[i+1])
				i += 1
			case 'n', 'N':
				buf = append(buf, '\n')
				i += 1
			case 'r', 'R':
				buf = append(buf, '\r')
				i += 1
			case 'v', 'V':
				buf = append(buf, '\v')
				i += 1
			case 't', 'T':
				buf = append(buf, '\t')
				i += 1
			case 'b', 'B':
				buf = append(buf, '\b')
				i += 1
			case 'f', 'F':
				buf = append(buf, '\f')
				i += 1

			case 'x', 'X':
				// Byte escape sequence
				// x12
				if i+3 >= length {
					return "", start, fmt.Errorf("parseQuotedName: Bad hex escape length: Pos=%v", i)
				}
				v, err := strconv.ParseInt(string(src[i+2:i+4]), 16, 64)
				if err != nil {
					return "", start, fmt.Errorf("parseQuotedName: Cannot parse hex unicode escape: Pos=%v", i)
				}
				buf = append(buf, rune(v))
				i += 3

			case 'u', 'U':
				// Unicode escape sequence
				if i+2 >= length {
					return "", start, fmt.Errorf("parseQuotedName: Bad unicode escape length (a): Pos=%v", i)
				}
				if src[i+2] == '{' {
					// u{1} , ... , u{123456}
					end, err := parseHex(src, i+3)
					if err != nil {
						return "", start, fmt.Errorf("parseQuotedName: Cannot parse unicode escape: Pos=%v", i)
					}
					if end > i+9 {
						return "", start, fmt.Errorf("parseQuotedName: Bad unicode escape length (b): Pos=%v", i)
					}

					v, err := strconv.ParseInt(string(src[i+3:end]), 16, 64)
					if err != nil {
						return "", start, err
					}
					buf = append(buf, rune(v))

					if end >= length {
						return "", start, fmt.Errorf("parseQuotedName: Unexpected termination on unicode escape (a): Pos=%v", i)
					}
					if src[end] != '}' {
						return "", start, fmt.Errorf("parseQuotedName: Unexpected termination on unicode escape (b): Pos=%v", i)
					}
					i = end // end is '}'

				} else {
					// u1234
					if i+5 >= length {
						return "", start, fmt.Errorf("parseQuotedName: Bad 4 digit unicode escape length (a): Pos=%v", i)
					}
					end, err := parseHex(src, i+2)
					if err != nil {
						return "", start, fmt.Errorf("parseQuotedName: Cannot parse 4 digit unicode escape: Pos=%v", i)
					}
					if end != i+6 {
						return "", start, fmt.Errorf("parseQuotedName: Bad 4 digit unicode escape length (b): Pos=%v", i)
					}

					v, err := strconv.ParseInt(string(src[i+2:end]), 16, 64)
					if err != nil {
						return "", start, err
					}
					buf = append(buf, rune(v))
					i = end - 1
				}
			}

		default:
			buf = append(buf, ch)
		}
	}
	return "", start, fmt.Errorf("parseQuotedName: Unexpected termination: Pos=%v", start)
}

func parseBareName(src []rune, start int) (string, int, error) {
	length := len(src)
	buf := make([]rune, 0, 32)
	var i int

	for i = start; i < length; i++ {
		ch := src[i]
		if unicode.IsSpace(ch) || unicode.IsControl(ch) {
			break
		}
		if '!' <= ch && ch <= '/' || ':' <= ch && ch <= '@' || '[' <= ch && ch <= '`' || '{' <= ch && ch <= '~' {
			break
		}
		buf = append(buf, src[i])
	}

	if len(buf) == 0 {
		return "", start, errors.New("parseBareName: Empty expression")
	}
	return string(buf), i, nil
}

func parseNumber(src []rune, start int) (int, error) {
	length := len(src)
	var i int

	for i = start; i < length; i++ {
		ch := src[i]
		if i == 0 && ch == '-' {
			continue
		}
		if '0' <= ch && ch <= '9' {
			continue
		}
		break
	}

	if i == start {
		return start, errors.New("parseNumber: Empty expression")
	}
	return i, nil
}

func parseHex(src []rune, start int) (int, error) {
	length := len(src)
	var i int

	for i = start; i < length; i++ {
		ch := src[i]
		if '0' <= ch && ch <= '9' || 'A' <= ch && ch <= 'F' || 'a' <= ch && ch <= 'f' {
			continue
		}
		break
	}

	if i == start {
		return start, errors.New("parseHex: Empty expression")
	}
	return i, nil
}
