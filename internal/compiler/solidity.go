// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package compiler wraps the Solidity and Vyper compiler executables (solc; vyper).
package compiler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func RichTypeToType(rich string) string {
	if strings.HasPrefix(rich, "t_") {
		rest := rich[2:]
		switch {
		case rest == "bool" || rest == "address" || rest == "string" || rest == "bytes":
			return rest
		case strings.HasPrefix(rest, "uint") || strings.HasPrefix(rest, "int"):
			return rest
		case strings.HasPrefix(rest, "ufixed") || strings.HasPrefix(rest, "fixed"):
			return rest
		case strings.HasPrefix(rest, "bytes"):
			return rest
		case strings.HasPrefix(rest, "enum"):
			return regexp.MustCompile(`t_enum\((.*)\)[0-9]+`).FindStringSubmatch(rich)[1]
		}
	}

	return rich
}

func GetSizeOfTypeIdentifier(id string) int {
	switch {
	case id == "t_string_storage_ptr":
		return 256
	case id == "t_bytes_storage_ptr":
		return 256
	case id == "t_bool":
		return 8
	case id == "t_address", id == "t_address_payable":
		return 160
	case strings.HasPrefix(id, "t_uint"):
		bitlen, err := strconv.Atoi(id[len("t_uint"):])
		if err != nil {
			panic(err)
		}
		return bitlen
	case strings.HasPrefix(id, "t_int"):
		bitlen, err := strconv.Atoi(id[len("t_int"):])
		if err != nil {
			panic(err)
		}
		return bitlen
	case strings.HasPrefix(id, "t_bytes"):
		bitlen, err := strconv.Atoi(id[len("t_bytes"):])
		if err != nil {
			panic(err)
		}
		return bitlen * 8
	}
	panic(id)
}

// Solidity contains information about the solidity compiler.
type Solidity struct {
	Path, Version, FullVersion string
	Major, Minor, Patch        int
}

// --combined-output format
type solcOutput struct {
	Contracts map[string]struct {
		BinRuntime                                  string `json:"bin-runtime"`
		SrcMapRuntime                               string `json:"srcmap-runtime"`
		Bin, SrcMap, Abi, Devdoc, Userdoc, Metadata string
		Hashes                                      map[string]string
	}
	Version string
}

// solidity v.0.8 changes the way ABI, Devdoc and Userdoc are serialized
type solcOutputV8 struct {
	Contracts map[string]struct {
		BinRuntime            string `json:"bin-runtime"`
		SrcMapRuntime         string `json:"srcmap-runtime"`
		Bin, SrcMap, Metadata string
		Abi                   interface{}
		Devdoc                interface{}
		Userdoc               interface{}
		Hashes                map[string]string
	}
	Version string
}

func (s *Solidity) makeArgs() []string {
	p := []string{
		"--combined-json", "bin,bin-runtime,srcmap,srcmap-runtime,abi,userdoc,devdoc",
		"--optimize",                  // code optimizer switched on
		"--allow-paths", "., ./, ../", // default to support relative paths
	}
	if s.Major > 0 || s.Minor > 4 || s.Patch > 6 {
		p[1] += ",metadata,hashes"
	}
	return p
}

// SolidityVersion runs solc and parses its version output.
func SolidityVersion(solc string) (*Solidity, error) {
	if solc == "" {
		solc = "solc"
	}
	var out bytes.Buffer
	cmd := exec.Command(solc, "--version")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	matches := versionRegexp.FindStringSubmatch(out.String())
	if len(matches) != 4 {
		return nil, fmt.Errorf("can't parse solc version %q", out.String())
	}
	s := &Solidity{Path: cmd.Path, FullVersion: out.String(), Version: matches[0]}
	if s.Major, err = strconv.Atoi(matches[1]); err != nil {
		return nil, err
	}
	if s.Minor, err = strconv.Atoi(matches[2]); err != nil {
		return nil, err
	}
	if s.Patch, err = strconv.Atoi(matches[3]); err != nil {
		return nil, err
	}
	return s, nil
}

// CompileSolidityString builds and returns all the contracts contained within a source string.
func CompileSolidityString(solc, source string) (map[string]*Contract, error) {
	if len(source) == 0 {
		return nil, errors.New("solc: empty source string")
	}
	s, err := SolidityVersion(solc)
	if err != nil {
		return nil, err
	}
	args := append(s.makeArgs(), "--")
	cmd := exec.Command(s.Path, append(args, "-")...)
	cmd.Stdin = strings.NewReader(source)
	return s.run(cmd, source)
}

// CompileSolidity compiles all given Solidity source files.
func CompileSolidity(solc string, sourcefiles ...string) (map[string]*Contract, error) {
	if len(sourcefiles) == 0 {
		return nil, errors.New("solc: no source files")
	}
	source, err := slurpFiles(sourcefiles)
	if err != nil {
		return nil, err
	}
	s, err := SolidityVersion(solc)
	if err != nil {
		return nil, err
	}
	args := append(s.makeArgs(), "--")
	cmd := exec.Command(s.Path, append(args, sourcefiles...)...)
	return s.run(cmd, source)
}

func (s *Solidity) run(cmd *exec.Cmd, source string) (map[string]*Contract, error) {
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("solc: %v\n%s", err, stderr.Bytes())
	}
	return ParseCombinedJSON(stdout.Bytes(), source, s.Version, s.Version, strings.Join(s.makeArgs(), " "))
}

// ParseCombinedJSON takes the direct output of a solc --combined-output run and
// parses it into a map of string contract name to Contract structs. The
// provided source, language and compiler version, and compiler options are all
// passed through into the Contract structs.
//
// The solc output is expected to contain ABI, source mapping, user docs, and dev docs.
//
// Returns an error if the JSON is malformed or missing data, or if the JSON
// embedded within the JSON is malformed.
func ParseCombinedJSON(combinedJSON []byte, source string, languageVersion string, compilerVersion string, compilerOptions string) (map[string]*Contract, error) {
	var output solcOutput
	if err := json.Unmarshal(combinedJSON, &output); err != nil {
		// Try to parse the output with the new solidity v.0.8.0 rules
		return parseCombinedJSONV8(combinedJSON, source, languageVersion, compilerVersion, compilerOptions)
	}
	// Compilation succeeded, assemble and return the contracts.
	contracts := make(map[string]*Contract)
	for name, info := range output.Contracts {
		// Parse the individual compilation results.
		var abi interface{}
		if err := json.Unmarshal([]byte(info.Abi), &abi); err != nil {
			return nil, fmt.Errorf("solc: error reading abi definition (%v)", err)
		}
		var userdoc, devdoc interface{}
		json.Unmarshal([]byte(info.Userdoc), &userdoc)
		json.Unmarshal([]byte(info.Devdoc), &devdoc)

		contracts[name] = &Contract{
			Code:        "0x" + info.Bin,
			RuntimeCode: "0x" + info.BinRuntime,
			Hashes:      info.Hashes,
			Info: ContractInfo{
				Source:          source,
				Language:        "Solidity",
				LanguageVersion: languageVersion,
				CompilerVersion: compilerVersion,
				CompilerOptions: compilerOptions,
				SrcMap:          info.SrcMap,
				SrcMapRuntime:   info.SrcMapRuntime,
				AbiDefinition:   abi,
				UserDoc:         userdoc,
				DeveloperDoc:    devdoc,
				Metadata:        info.Metadata,
			},
		}
	}
	return contracts, nil
}

// parseCombinedJSONV8 parses the direct output of solc --combined-output
// and parses it using the rules from solidity v.0.8.0 and later.
func parseCombinedJSONV8(combinedJSON []byte, source string, languageVersion string, compilerVersion string, compilerOptions string) (map[string]*Contract, error) {
	var output solcOutputV8
	if err := json.Unmarshal(combinedJSON, &output); err != nil {
		return nil, err
	}
	// Compilation succeeded, assemble and return the contracts.
	contracts := make(map[string]*Contract)
	for name, info := range output.Contracts {
		contracts[name] = &Contract{
			Code:        "0x" + info.Bin,
			RuntimeCode: "0x" + info.BinRuntime,
			Hashes:      info.Hashes,
			Info: ContractInfo{
				Source:          source,
				Language:        "Solidity",
				LanguageVersion: languageVersion,
				CompilerVersion: compilerVersion,
				CompilerOptions: compilerOptions,
				SrcMap:          info.SrcMap,
				SrcMapRuntime:   info.SrcMapRuntime,
				AbiDefinition:   info.Abi,
				UserDoc:         info.Userdoc,
				DeveloperDoc:    info.Devdoc,
				Metadata:        info.Metadata,
			},
		}
	}
	return contracts, nil
}

type SolidityCompiler struct {
	version string
	path    string
}

func NewSolidityCompiler(version string) (Compiler, error) {
	compilerDir := path.Join(os.TempDir(), "solidity", version)

	var compilerPath string
	if runtime.GOOS == "windows" {
		compilerPath = path.Join(compilerDir, "solc.exe")
	} else {
		compilerPath = path.Join(compilerDir, "solc-static-linux")
	}

	if _, err := os.Stat(compilerPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to check compiler: %w", err)
		}
		err = downloadSolidityCompiler(version, compilerDir)
		if err != nil {
			return nil, fmt.Errorf("failed to download compiler: %w", err)
		}
	}

	return &SolidityCompiler{
		version: version,
		path:    compilerPath,
	}, nil
}

func downloadSolidityCompiler(version string, dest string) error {
	err := os.MkdirAll(dest, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("failed to create parent dir: %w", err)
	}

	var url string
	if runtime.GOOS == "linux" {
		url = fmt.Sprintf("https://github.com/ethereum/solidity/releases/download/v%s/solc-static-linux", version)
	} else if runtime.GOOS == "windows" {
		url = fmt.Sprintf("https://github.com/ethereum/solidity/releases/download/v%s/solidity-windows.zip", version)
	} else {
		return fmt.Errorf("unsupported os %s", runtime.GOOS)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download solidity v%s: %w", version, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if runtime.GOOS == "linux" {
		err = ioutil.WriteFile(path.Join(dest, path.Base(url)), body, os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("failed to write compiler: %w", err)
		}
	} else {
		reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}

		for _, file := range reader.File {
			if strings.Contains(file.Name, "..") || strings.HasPrefix(file.Name, "/") {
				return fmt.Errorf("unsafe path: %s", file.Name)
			}

			reader, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to unpack %s: %w", file.Name, err)
			}
			body, err = ioutil.ReadAll(reader)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", file.Name, err)
			}
			err = ioutil.WriteFile(path.Join(dest, file.Name), body, file.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to write %s: %w", file.Name, err)
			}
			err = reader.Close()
			if err != nil {
				return fmt.Errorf("failed to close %s: %w", file.Name, err)
			}
		}
	}

	return nil
}

func (c *SolidityCompiler) CompileFromString(src string) (map[string]*Contract, error) {
	contracts, err := CompileSolidityString(c.path, src)
	if err != nil {
		return nil, err
	}

	remappedContracts := make(map[string]*Contract)
	for k, v := range contracts {
		remappedContracts[strings.TrimPrefix(k, "<stdin>:")] = v
	}
	return remappedContracts, nil
}

type StorageEntry struct {
	Label    string `json:"label"`
	AstID    int    `json:"astId"`
	Contract string `json:"contract"`
	Slot     string `json:"slot"`
	Offset   int    `json:"offset"`
	Type     string `json:"type"`
}

type Encoding string

const (
	EncodingInplace      Encoding = "inplace"
	EncodingMapping               = "mapping"
	EncodingBytes                 = "bytes"
	EncodingDynamicArray          = "dynamic_array"
)

type StorageType struct {
	Label         string   `json:"string"`
	NumberOfBytes string   `json:"numberOfBytes"`
	Encoding      Encoding `json:"encoding"`
	// only set if is struct
	Members []*StorageEntry `json:"members"`
	// only set if mapping
	Key   string `json:"key"`
	Value string `json:"value"`
	// only set if array
	Base string `json:"base"`
}

type StorageLayout struct {
	Storage []*StorageEntry         `json:"storage"`
	Types   map[string]*StorageType `json:"types"`
}

type StandardJsonContract struct {
	ABI           any            `json:"abi"`
	StorageLayout *StorageLayout `json:"storageLayout"`
}

type LegacyASTNode struct {
	ID         int              `json:"id"`
	Name       string           `json:"name"`
	Src        string           `json:"src"`
	Attributes map[string]bool  `json:"attributes"`
	Children   []*LegacyASTNode `json:"children"`
}

type ASTNode struct {
	ID       int    `json:"id"`
	NodeType string `json:"nodeType"`
	Node     any    `json:"-"`
}

type EnumValue struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EnumDefinitionNode struct {
	ID            int          `json:"id"`
	Members       []*EnumValue `json:"members"`
	CanonicalName string       `json:"canonicalName"`
	Name          string       `json:"name"`
}

type StructDefinitionNode struct {
	ID            int                        `json:"id"`
	CanonicalName string                     `json:"canonicalName"`
	Members       []*VariableDeclarationNode `json:"members"`
	Name          string                     `json:"name"`
}

func (n *ASTNode) UnmarshalJSON(b []byte) error {
	getType := struct {
		ID       int    `json:"id"`
		NodeType string `json:"nodeType"`
	}{}
	if err := json.Unmarshal(b, &getType); err != nil {
		return err
	}

	n.ID = getType.ID
	n.NodeType = getType.NodeType

	switch getType.NodeType {
	case "ContractDefinition":
		n.Node = &ContractDefinitionNode{}
	case "PragmaDirective":
		n.Node = &PragmaDirectiveNode{}
	case "VariableDeclaration":
		n.Node = &VariableDeclarationNode{}
	case "EnumDefinition":
		n.Node = &EnumDefinitionNode{}
	case "StructDefinition":
		n.Node = &StructDefinitionNode{}
	case "Literal":
		n.Node = &LiteralNode{}
	default:
		//return fmt.Errorf("unhandled node %v\n", getType.NodeType)
	}

	return json.Unmarshal(b, &n.Node)
}

type LiteralNode struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type PragmaDirectiveNode struct {
}

type ContractDefinitionNode struct {
	ID                      int        `json:"id"`
	LinearizedBaseContracts []int      `json:"linearizedBaseContracts"`
	Name                    string     `json:"name"`
	Nodes                   []*ASTNode `json:"nodes"`
}

type TypeDescriptions struct {
	TypeIdentifier string `json:"typeIdentifier"`
	TypeString     string `json:"typeString"`
}

type TypeName struct {
	ID               int               `json:"id"`
	TypeDescriptions *TypeDescriptions `json:"typeDescriptions"`
	NodeType         string            `json:"nodeType"`

	// only for array types
	BaseType *TypeName `json:"baseType,omitempty"`
	Length   *ASTNode  `json:"length,omitempty"`

	// only for function types
	ParameterTypes       any `json:"parameterTypes,omitempty"`
	ReturnParameterTypes any `json:"returnParameterTypes,omitempty"`

	// only for mapping types
	KeyType   *TypeName `json:"keyType,omitempty"`
	ValueType *TypeName `json:"valueType,omitempty"`

	// only for user defined types
	ReferencedDeclaration *int `json:"referencedDeclaration,omitempty"`
}

func (t *TypeName) IsArray() bool {
	return t.NodeType == "ArrayTypeName"
}

func (t *TypeName) IsFunction() bool {
	return t.NodeType == "FunctionTypeName"
}

func (t *TypeName) IsMapping() bool {
	return t.NodeType == "Mapping"
}

func (t *TypeName) IsUserDefinedType() bool {
	return t.NodeType == "UserDefinedTypeName"
}

func (t *TypeName) IsElementaryType() bool {
	return t.NodeType == "ElementaryTypeName"
}

type VariableDeclarationNode struct {
	Name             string            `json:"name"`
	TypeDescriptions *TypeDescriptions `json:"typeDescriptions"`
	Constant         bool              `json:"constant"`
	Mutability       string            `json:"mutability"`

	TypeName *TypeName `json:"typeName"`
}

type SourceUnitNode struct {
	AbsolutePath    string           `json:"absolutePath"`
	ExportedSymbols map[string][]int `json:"exportedSymbols"`
	ID              int              `json:"id"`
	NodeType        string           `json:"nodeType"`
	Src             string           `json:"src"`
	Nodes           []*ASTNode       `json:"nodes"`
}

type StandardJsonSource struct {
	ID        int             `json:"id"`
	AST       *SourceUnitNode `json:"ast"`
	LegacyAST *LegacyASTNode  `json:"legacyAST"`
}

type StandardJsonError struct {
	SourceLocation   map[string]any `json:"sourceLocation"`
	Type             string         `json:"type"`
	Severity         string         `json:"severity"`
	ErrorCode        string         `json:"errorCode"`
	Message          string         `json:"message"`
	FormattedMessage string         `json:"formattedMessage"`
}

type StandardJsonOutput struct {
	Errors    []*StandardJsonError                        `json:"errors"`
	Contracts map[string]map[string]*StandardJsonContract `json:"contracts"`
	Sources   map[string]*StandardJsonSource              `json:"sources"`
}

type StandardJsonSourceFile struct {
	Keccak256 string   `json:"keccak256"`
	URLs      []string `json:"urls"`
	Content   string   `json:"content"`
}

type StandardJsonInput struct {
	Language string                             `json:"language"`
	Sources  map[string]*StandardJsonSourceFile `json:"sources"`
	Settings map[string]any                     `json:"settings"`
}

func (c *SolidityCompiler) CompileFromStandardJSON(input *StandardJsonInput) (*StandardJsonOutput, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	s, err := SolidityVersion(c.path)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(s.Path, "--standard-json")
	cmd.Stdin = bytes.NewReader(b)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("solc: %v\n%s", err, stderr.Bytes())
	}
	//fmt.Println(string(stdout.Bytes()))

	os.WriteFile("/tmp/out.json", stdout.Bytes(), os.FileMode(0755))

	var out StandardJsonOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return &out, nil
}

func ExtractCodeAndABI(contract *Contract) ([]byte, *abi.ABI, error) {
	code, err := hexutil.Decode(contract.RuntimeCode)
	if err != nil {
		return nil, nil, fmt.Errorf("expected hex runtimecode: %w", err)
	}

	encoded, err := json.Marshal(contract.Info.AbiDefinition)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal abi: %w", err)
	}

	abi, err := abi.JSON(bytes.NewReader(encoded))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse abi: %w", err)
	}

	return code, &abi, nil
}
