package solidity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	cache     = make(map[*abi.Event][]abi.Arguments)
	cacheLock = sync.RWMutex{}
)

func DecodeParameters(event *abi.Event, log *types.Log) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	args := loadArgumentsFromCache(event)

	if err := abi.ParseTopicsIntoMap(result, args[0], log.Topics[1:]); err != nil {
		return nil, fmt.Errorf("failed to parse topics: %w", err)
	}

	if err := event.Inputs.UnpackIntoMap(result, log.Data); err != nil {
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}

	return result, nil
}

func loadArgumentsFromCache(event *abi.Event) []abi.Arguments {
	cacheLock.RLock()
	result := cache[event]
	cacheLock.RUnlock()

	if result != nil {
		return result
	}

	var indexedInputs abi.Arguments
	var nonIndexedInputs abi.Arguments

	for _, input := range event.Inputs {
		if input.Indexed {
			indexedInputs = append(indexedInputs, input)
		} else {
			nonIndexedInputs = append(nonIndexedInputs, input)
		}
	}

	result = []abi.Arguments{indexedInputs, nonIndexedInputs}

	cacheLock.Lock()
	cache[event] = result
	cacheLock.Unlock()

	return result
}

func MustDecodeEventSignature(sig string) *abi.Event {
	event, err := DecodeEventSignature(sig)
	if err != nil {
		panic(err)
	}
	return event
}

func MustDecodeErrorSignature(sig string) *abi.Error {
	event, err := DecodeErrorSignature(sig)
	if err != nil {
		panic(err)
	}
	return event
}

func DecodeFunctionSignature(sig string) (*abi.Method, error) {
	if strings.HasPrefix(sig, "function ") {
		sig = sig[9:]
	}

	start := strings.Index(sig, "(")
	if start == -1 {
		return nil, fmt.Errorf("could not find name")
	}

	name := sig[:start]
	params := sig[start:]

	if findClosingBracket(params, '(', ')') != len(params)-1 {
		return nil, fmt.Errorf("signature is not balanced")
	}

	args, err := rewriteTuple(params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal([]struct {
		Type    string
		Name    string
		Inputs  []abi.ArgumentMarshaling
		Outputs []abi.ArgumentMarshaling

		// Status indicator which can be: "pure", "view",
		// "nonpayable" or "payable".
		StateMutability string

		// Deprecated Status indicators, but removed in v0.6.0.
		Constant bool // True if function is either pure or view
		Payable  bool // True if function is payable

		// Event relevant indicator represents the event is
		// declared as anonymous.
		Anonymous bool
	}{
		{
			Type:   "function",
			Name:   name,
			Inputs: args,
		},
	})
	if err != nil {
		return nil, err
	}

	v, err := abi.JSON(bytes.NewReader(b))
	if err != nil {
		panic(fmt.Errorf("failed to decode %s: %w", sig, err))
	}

	e := v.Methods[name]
	return &e, nil
}

func DecodeErrorSignature(sig string) (*abi.Error, error) {
	if strings.HasPrefix(sig, "error ") {
		sig = sig[6:]
	}

	start := strings.Index(sig, "(")
	if start == -1 {
		return nil, fmt.Errorf("could not find name")
	}

	name := sig[:start]
	params := sig[start:]

	if findClosingBracket(params, '(', ')') != len(params)-1 {
		return nil, fmt.Errorf("signature is not balanced")
	}

	args, err := rewriteTuple(params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal([]struct {
		Type    string
		Name    string
		Inputs  []abi.ArgumentMarshaling
		Outputs []abi.ArgumentMarshaling

		// Status indicator which can be: "pure", "view",
		// "nonpayable" or "payable".
		StateMutability string

		// Deprecated Status indicators, but removed in v0.6.0.
		Constant bool // True if function is either pure or view
		Payable  bool // True if function is payable

		// Event relevant indicator represents the event is
		// declared as anonymous.
		Anonymous bool
	}{
		{
			Type:   "error",
			Name:   name,
			Inputs: args,
		},
	})
	if err != nil {
		return nil, err
	}

	v, err := abi.JSON(bytes.NewReader(b))
	if err != nil {
		panic(fmt.Errorf("failed to decode %s: %w", sig, err))
	}

	e := v.Errors[name]
	return &e, nil
}

func DecodeEventSignature(sig string) (*abi.Event, error) {
	if strings.HasPrefix(sig, "event ") {
		sig = sig[6:]
	}

	start := strings.Index(sig, "(")
	if start == -1 {
		return nil, fmt.Errorf("could not find name")
	}

	name := sig[:start]
	params := sig[start:]

	if findClosingBracket(params, '(', ')') != len(params)-1 {
		return nil, fmt.Errorf("signature is not balanced")
	}

	args, err := rewriteTuple(params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal([]struct {
		Type    string
		Name    string
		Inputs  []abi.ArgumentMarshaling
		Outputs []abi.ArgumentMarshaling

		// Status indicator which can be: "pure", "view",
		// "nonpayable" or "payable".
		StateMutability string

		// Deprecated Status indicators, but removed in v0.6.0.
		Constant bool // True if function is either pure or view
		Payable  bool // True if function is payable

		// Event relevant indicator represents the event is
		// declared as anonymous.
		Anonymous bool
	}{
		{
			Type:      "event",
			Name:      name,
			Inputs:    args,
			Anonymous: name == "",
		},
	})
	if err != nil {
		return nil, err
	}

	v, err := abi.JSON(bytes.NewReader(b))
	if err != nil {
		panic(fmt.Errorf("failed to decode %s: %w", sig, err))
	}

	e := v.Events[name]
	return &e, nil
}

func rewriteTuple(params string) ([]abi.ArgumentMarshaling, error) {
	if params[0] != '(' {
		return nil, fmt.Errorf("expected open bracket")
	}
	if params[len(params)-1] != ')' {
		return nil, fmt.Errorf("expected close bracket")
	}

	var args []abi.ArgumentMarshaling

	params = params[1 : len(params)-1]
	for len(params) > 0 {
		params = strings.TrimLeft(params, " ")
		var rewritten abi.ArgumentMarshaling

		if params[0] == '(' {
			end := findClosingBracket(params, '(', ')')

			args, err := rewriteTuple(params[:end+1])
			if err != nil {
				return nil, err
			}

			params = params[end+1:]

			nextComma := strings.Index(params, ",")

			var variable string
			if nextComma != -1 {
				variable = params[:nextComma]
				params = params[nextComma+1:]
			} else {
				variable = params
				params = ""
			}
			variable = strings.TrimSpace(variable)

			parts := strings.Split(variable, " ")
			switch len(parts) {
			case 1:
				rewritten = abi.ArgumentMarshaling{
					Name:       parts[0],
					Type:       "tuple",
					Components: args,
				}
			case 2:
				rewritten = abi.ArgumentMarshaling{
					Name:       parts[1],
					Type:       "tuple",
					Components: args,
					Indexed:    true,
				}
			}
		} else {
			nextComma := strings.Index(params, ",")

			var variable string
			if nextComma != -1 {
				variable = params[:nextComma]
				params = params[nextComma+1:]
			} else {
				variable = params
				params = ""
			}
			variable = strings.TrimSpace(variable)

			parts := strings.Split(variable, " ")

			switch len(parts) {
			case 1:
				rewritten = abi.ArgumentMarshaling{
					Type: parts[0],
				}
			case 2:
				rewritten = abi.ArgumentMarshaling{
					Type: parts[0],
					Name: parts[1],
				}
			case 3:
				rewritten = abi.ArgumentMarshaling{
					Type:    parts[0],
					Name:    parts[2],
					Indexed: true,
				}
			}
		}

		if rewritten.Type == "uint" {
			rewritten.Type = "uint256"
		} else if rewritten.Type == "int" {
			rewritten.Type = "int256"
		}

		args = append(args, rewritten)
	}

	for i := range args {
		if args[i].Name == "" {
			args[i].Name = fmt.Sprintf("arg%d", i)
		}
	}

	return args, nil
}

func findClosingBracket(in string, openCh rune, closeCh rune) int {
	if []rune(in)[0] != openCh {
		return -1
	}

	depth := 0
	for i, c := range in {
		if c == openCh {
			depth++
		} else if c == closeCh {
			depth--

			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

var signatureRe = regexp.MustCompile(`^([a-zA-Z$_][a-zA-Z0-9$_]*)(\(.*\))$`)

func VerifySignature(sig string) bool {
	res := signatureRe.FindStringSubmatch(sig)
	if res == nil {
		return false
	}

	return check(res[2])
}

func check(sig string) bool {
	if len(sig) == 0 {
		return false
	}

	if sig[len(sig)-1] == ']' {
		return checkArrayRecursively(sig)
	}

	if sig[0] == '(' && sig[len(sig)-1] == ')' {
		return checkTupleRecursively(sig)
	}

	return checkType(sig)
}

func checkTupleRecursively(tuple string) bool {
	if tuple[0] != '(' || tuple[len(tuple)-1] != ')' {
		return false
	}

	tupleContents := tuple[1 : len(tuple)-1]

	for len(tupleContents) > 0 {
		var endOfFirstElement int
		if tupleContents[0] == '(' {
			// first element is another tuple, find the balanced closing bracket
			closingBracketIndex := findClosingBracket(tupleContents, '(', ')')
			if closingBracketIndex == -1 {
				return false
			}
			endOfFirstElement = closingBracketIndex + 1
		} else {
			// first element is a basic type, look for a comma to determine if there's another element or not
			commaIndex := strings.Index(tupleContents, ",")
			if commaIndex == -1 {
				// no comma, there is only one element
				endOfFirstElement = len(tupleContents)
			} else {
				// found a comma, that's the end of the current element
				endOfFirstElement = commaIndex
			}
		}

		for endOfFirstElement < len(tupleContents) && tupleContents[endOfFirstElement] == '[' {
			// if we find something that looks like it might be an array, greedily consume it. we'll verify it later
			closingArrayIndex := findClosingBracket(tupleContents[endOfFirstElement:], '[', ']')
			if closingArrayIndex == -1 {
				return false
			}
			endOfFirstElement += closingArrayIndex + 1
		}

		firstElement := tupleContents[:endOfFirstElement]
		tupleContents = tupleContents[endOfFirstElement:]
		if !check(firstElement) {
			return false
		}

		// now if there's a comma remaining, there should be another element
		if len(tupleContents) > 0 && tupleContents[0] == ',' {
			tupleContents = tupleContents[1:]

			// there should be something after the comma
			if len(tupleContents) == 0 {
				return false
			}
		}
	}

	return true
}

func checkArrayRecursively(sig string) bool {
	if sig[len(sig)-1] != ']' {
		return false
	}

	if strings.HasSuffix(sig, "[]") {
		return check(sig[:len(sig)-2])
	}

	arrayStartIndex := strings.LastIndexByte(sig, '[')
	if arrayStartIndex == -1 {
		return false
	}

	arrLenStr := sig[arrayStartIndex+1 : len(sig)-1]

	arrLen, err := strconv.ParseUint(arrLenStr, 10, 64)
	if err != nil || arrLen == 0 {
		return false
	}

	return check(sig[:arrayStartIndex])
}

func checkType(typ string) bool {
	switch {
	case strings.HasPrefix(typ, "uint"):
		return checkLength(strings.TrimPrefix(typ, "uint"), 0, 256, 8)
	case strings.HasPrefix(typ, "int"):
		return checkLength(strings.TrimPrefix(typ, "int"), 0, 256, 8)
	case strings.HasPrefix(typ, "bytes"):
		return typ == "bytes" || checkLength(strings.TrimPrefix(typ, "bytes"), 0, 32, 1)
	}

	switch typ {
	case "address", "bool", "function", "bytes", "string":
		return true
	}

	return false
}

func checkLength(strVal string, min uint64, max uint64, mod uint64) bool {
	val, err := strconv.ParseUint(strVal, 10, 64)
	if err != nil {
		return false
	}
	if val <= min || val > max {
		return false
	}
	if val%mod != 0 {
		return false
	}
	return true
}
