package compiler

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math"
	"math/big"
	"strconv"
)

// https://solidity-ast.netlify.app/

type ASTVariable struct {
	// the name of this variable as defined in the contract code itself
	Name string `json:"name"`

	// the full name of the part of the variable at this specific slot and offset
	// for primitives it will be equal to name
	// for arrays/structs it will contain the field name or array offset
	FullName string `json:"fullName"`

	TypeName *TypeName `json:"typeName"`

	Bits int `json:"bits"`
}

func (v *ASTVariable) String() string {
	return fmt.Sprintf("ASTVariable{Name=%s,FullName=%s,TypeName=%s,TypeIdentifier=%s}", v.Name, v.FullName, v.TypeName.TypeDescriptions.TypeString, v.TypeName.TypeDescriptions.TypeIdentifier)
}

type ASTStorageLayout struct {
	// contains the exact variable at each slot and offset
	// structs and fixed size arrays are resolved, to find the original see Arrays and Structs
	Slots map[common.Hash]map[int]*ASTVariable `json:"slots"`

	// contains the slot at which each fixed-size array starts
	Arrays map[common.Hash]*ASTVariable `json:"arrays"`

	// contains the slot at which each fixed-size struct starts
	Structs map[common.Hash]*ASTVariable `json:"structs"`

	// stores information about structs
	AllStructs map[string]*ASTStorageLayout `json:"allStructs"`
}

func GenerateStorageLayout(nodesById map[int]*ASTNode, vars []*VariableDeclarationNode) *ASTStorageLayout {
	var astVars []*ASTVariable
	for _, vdn := range vars {
		astVars = append(astVars, &ASTVariable{
			Name:     vdn.Name,
			FullName: vdn.Name,
			TypeName: vdn.TypeName,
		})
	}

	return generateStructLayout(nodesById, astVars)
}

func generateStructLayout(nodesById map[int]*ASTNode, types []*ASTVariable) *ASTStorageLayout {
	result := &ASTStorageLayout{
		Slots:      make(map[common.Hash]map[int]*ASTVariable),
		Arrays:     make(map[common.Hash]*ASTVariable),
		Structs:    make(map[common.Hash]*ASTVariable),
		AllStructs: make(map[string]*ASTStorageLayout),
	}

	var (
		currentSlot   int
		currentOffset int
	)
	for _, astVar := range types {
		currentSlot, currentOffset = result.assignSlot(nodesById, astVar, currentSlot, currentOffset)
	}

	return result
}

func (l *ASTStorageLayout) putInSlot(slot int, offset int, astVar *ASTVariable) {
	slotHash := common.BigToHash(big.NewInt(int64(slot)))
	if _, ok := l.Slots[slotHash]; !ok {
		l.Slots[slotHash] = make(map[int]*ASTVariable)
	}
	l.Slots[slotHash][offset] = astVar
}

func (l *ASTStorageLayout) registerMeta(nodesById map[int]*ASTNode, typeName *TypeName) {
	if typeName.IsUserDefinedType() {
		node := nodesById[*typeName.ReferencedDeclaration]
		if structDef, ok := node.Node.(*StructDefinitionNode); ok {
			if _, ok := l.AllStructs[structDef.CanonicalName]; !ok {
				var astVars []*ASTVariable
				for _, member := range structDef.Members {
					astVars = append(astVars, &ASTVariable{
						Name:     member.Name,
						FullName: member.Name,
						TypeName: member.TypeName,
					})
				}
				resultingLayout := generateStructLayout(nodesById, astVars)
				for k, v := range resultingLayout.AllStructs {
					l.AllStructs[k] = v
				}
				resultingLayout.AllStructs = make(map[string]*ASTStorageLayout)
				l.AllStructs[structDef.CanonicalName] = resultingLayout
			}
		}
	} else if typeName.IsMapping() {
		l.registerMeta(nodesById, typeName.KeyType)
		l.registerMeta(nodesById, typeName.ValueType)
	} else if typeName.IsArray() {
		l.registerMeta(nodesById, typeName.BaseType)
	}
}

func (l *ASTStorageLayout) assignSlot(nodesById map[int]*ASTNode, astVar *ASTVariable, currentSlot int, currentOffset int) (int, int) {
	typeName := astVar.TypeName
	l.registerMeta(nodesById, typeName)

	if typeName.IsMapping() {

		if currentOffset > 0 {
			currentSlot += 1
			currentOffset = 0
		}

		l.putInSlot(currentSlot, 0, astVar)
		currentSlot += 1
	} else if typeName.IsArray() {
		if typeName.Length == nil {
			// dynamic array gets put at the keccak
			if currentOffset > 0 {
				currentSlot += 1
				currentOffset = 0
			}

			l.Arrays[common.BigToHash(big.NewInt(int64(currentSlot)))] = astVar
			l.putInSlot(currentSlot, 0, &ASTVariable{
				Name:     astVar.Name,
				FullName: fmt.Sprintf("%s.length", astVar.Name),
				TypeName: &TypeName{
					ID: -1,
					TypeDescriptions: &TypeDescriptions{
						TypeIdentifier: "t_uint256",
						TypeString:     "uint256",
					},
					NodeType: "ElementaryTypeName",
				},
				Bits: 256,
			})
			currentSlot += 1
		} else {
			// fixed size array gets inlined
			if currentOffset > 0 {
				currentOffset = 0
				currentSlot += 1
			}

			lenStr := typeName.Length.Node.(*LiteralNode).Value
			lenVal, err := strconv.Atoi(lenStr)
			if err != nil {
				panic(err)
			}

			l.Arrays[common.BigToHash(big.NewInt(int64(currentSlot)))] = astVar

			for i := 0; i < lenVal; i++ {
				currentSlot, currentOffset = l.assignSlot(nodesById, &ASTVariable{
					Name:     astVar.Name,
					FullName: fmt.Sprintf("%s[%d]", astVar.FullName, i),
					TypeName: typeName.BaseType,
				}, currentSlot, currentOffset)
			}

			if currentOffset > 0 {
				currentSlot += 1
				currentOffset = 0
			}
		}
	} else if typeName.IsUserDefinedType() {
		node := nodesById[*typeName.ReferencedDeclaration]
		switch v := node.Node.(type) {
		case *EnumDefinitionNode:
			numBits := int(math.Floor(math.Log2(float64(len(v.Members))))) + 1
			numBytes := int(math.Ceil(float64(numBits) / 8))
			size := numBytes * 8
			astVar.Bits = size

			if size+currentOffset <= 256 {
				l.putInSlot(currentSlot, currentOffset, astVar)
				currentOffset += size
			} else {
				if currentOffset > 0 {
					currentSlot += 1
					currentOffset = 0
				}

				l.putInSlot(currentSlot, 0, astVar)
				currentOffset = size
			}
		case *ContractDefinitionNode:
			size := 160
			astVar.Bits = 160

			if size+currentOffset <= 256 {
				l.putInSlot(currentSlot, currentOffset, astVar)
				currentOffset += size
			} else {
				if currentOffset > 0 {
					currentSlot += 1
					currentOffset = 0
				}

				l.putInSlot(currentSlot, 0, astVar)
				currentOffset = 160
			}
		case *StructDefinitionNode:
			// struct gets inlined
			if currentOffset > 0 {
				currentOffset = 0
				currentSlot += 1
			}

			l.Structs[common.BigToHash(big.NewInt(int64(currentSlot)))] = astVar

			for _, member := range v.Members {
				currentSlot, currentOffset = l.assignSlot(nodesById, &ASTVariable{
					Name:     astVar.Name,
					FullName: fmt.Sprintf("%s.%s", astVar.FullName, member.Name),
					TypeName: member.TypeName,
				}, currentSlot, currentOffset)
			}

			if currentOffset > 0 {
				currentSlot += 1
				currentOffset = 0
			}
		default:
			panic(fmt.Sprintf("unsupported type %+v", node))
		}
	} else if typeName.IsElementaryType() {
		if typeName.TypeDescriptions.TypeIdentifier == "t_string_storage_ptr" {
			// string pointer maybe gets put at the keccak, maybe not
			if currentOffset > 0 {
				currentSlot += 1
				currentOffset = 0
			}

			l.Arrays[common.BigToHash(big.NewInt(int64(currentSlot)))] = astVar
			l.putInSlot(currentSlot, 0, &ASTVariable{
				Name:     astVar.Name,
				FullName: astVar.Name,
				TypeName: &TypeName{
					ID: -1,
					TypeDescriptions: &TypeDescriptions{
						TypeIdentifier: "t_string_header",
						TypeString:     "stringHeader",
					},
					NodeType: "ElementaryTypeName",
				},
				Bits: 256,
			})
			currentSlot += 1
		} else {
			size := GetSizeOfTypeIdentifier(typeName.TypeDescriptions.TypeIdentifier)
			fmt.Printf("assigning %+v to %d %d (%d)\n", typeName, currentSlot, currentOffset, size)
			astVar.Bits = size
			if size+currentOffset <= 256 {
				l.putInSlot(currentSlot, currentOffset, astVar)
				currentOffset += size
			} else {
				if currentOffset > 0 {
					currentSlot += 1
					currentOffset = 0
				}

				l.putInSlot(currentSlot, 0, astVar)

				if size == 256 {
					currentSlot += 1
				} else {
					currentOffset = size
				}
			}
		}
	}

	return currentSlot, currentOffset
}
