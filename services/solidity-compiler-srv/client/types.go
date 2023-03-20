package solidityclient

import "errors"

var ErrCompilerUnavailable = errors.New("compiler unavailable")

type SolcSource struct {
	Content string `json:"content"`
}

type SolcStandardInput struct {
	Language string                `json:"language"`
	Sources  map[string]SolcSource `json:"sources"`
	Settings map[string]any        `json:"settings"`
}

type SolcStandardOutput struct {
	Errors    []SolcError                         `json:"errors"`
	Contracts map[string]map[string]*SolcContract `json:"contracts"`
}

type SolcError struct {
	//Component        string `json:"component"`
	//FormattedMessage string `json:"formattedMessage"`
	//Message          string `json:"message"`
	Severity string `json:"severity"`
	//SourceLocation   string `json:"sourceLocation"`
	//Type             string `json:"type"`
}

type SolcContract struct {
	ABI      any     `json:"abi"`
	Metadata string  `json:"metadata"`
	EVM      SolcEVM `json:"evm"`
}

type SolcEVM struct {
	Bytecode         SolcBytecode `json:"bytecode"`
	DeployedBytecode SolcBytecode `json:"deployedBytecode"`
}

type SolcBytecode struct {
	Object         string                                    `json:"object"`
	SourceMap      string                                    `json:"sourceMap"`
	LinkReferences map[string]map[string][]SolcLinkReference `json:"linkReferences"`
}

type SolcLinkReference struct {
	Start  uint64 `json:"start"`
	Length uint64 `json:"length"`
}

type CompileRequest struct {
	Version string             `json:"version"`
	Input   *SolcStandardInput `json:"input"`
}

type CompileResponse struct {
	Ok     bool                `json:"ok"`
	Error  string              `json:"error"`
	Result *SolcStandardOutput `json:"result"`
}
