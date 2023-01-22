package client

type SignatureType string

const (
	SignatureTypeFunction SignatureType = "function"
	SignatureTypeEvent                  = "event"
)

func SignatureTypes() []SignatureType {
	return []SignatureType{SignatureTypeFunction, SignatureTypeEvent}
}

func (t SignatureType) Valid() bool {
	return t == SignatureTypeFunction || t == SignatureTypeEvent
}

type AllTypes[T any] map[SignatureType]T

type ImportRequest AllTypes[[]string]

type ImportResponse AllTypes[*ImportResponseDetails]

type ImportResponseDetails struct {
	Imported   map[string]string `json:"imported"`
	Duplicated map[string]string `json:"duplicated"`
	Invalid    []string          `json:"invalid"`
}

func NewImportResponse() ImportResponse {
	response := make(ImportResponse)
	for _, typ := range SignatureTypes() {
		response[typ] = NewImportResponseDetails()
	}
	return response
}

func NewImportResponseDetails() *ImportResponseDetails {
	return &ImportResponseDetails{
		Imported:   make(map[string]string),
		Duplicated: make(map[string]string),
		Invalid:    nil,
	}
}

type SignatureData struct {
	Name     string `json:"name"`
	Filtered bool   `json:"filtered"`
}

type SignatureResponse AllTypes[map[string][]*SignatureData]

func NewSignatureResponse() SignatureResponse {
	response := make(SignatureResponse)
	for _, typ := range SignatureTypes() {
		response[typ] = make(map[string][]*SignatureData)
	}
	return response
}

type StatsResponse struct {
	Count AllTypes[int] `json:"count"`
}

func NewStatsResponse() *StatsResponse {
	return &StatsResponse{
		Count: make(AllTypes[int]),
	}
}
