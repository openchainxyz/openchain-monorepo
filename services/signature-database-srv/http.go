package signature_database_srv

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/core"
	"github.com/openchainxyz/openchainxyz-monorepo/services/signature-database-srv/client"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strings"
)

func fail(w http.ResponseWriter, status int, err error, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	log.WithError(err).Errorf(msg)
	json.NewEncoder(w).Encode(map[string]any{
		"ok":    false,
		"error": msg,
	})
}

func succeed(w http.ResponseWriter, result any) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"result": result,
	})
}

func (s *Service) serveLookup(w http.ResponseWriter, r *http.Request) {
	var err error

	response := client.NewSignatureResponse()

	params := r.URL.Query()
	shouldFilter := !params.Has("filter") || params.Get("filter") != "false"

	for _, typ := range client.SignatureTypes() {
		data := params.Get(string(typ))
		if len(data) == 0 {
			continue
		}
		response[typ], err = s.db.LoadSignatures(typ, strings.Split(data, ","))
		if err != nil {
			fail(w, http.StatusInternalServerError, err, "failed to load signatures")
			return
		}
	}

	s.filterResponse(response, shouldFilter)
	s.logSignatureResponse(r, response)

	succeed(w, response)
}

func (s *Service) serveSearch(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	query := params.Get("query")
	shouldFilter := !params.Has("filter") || params.Get("filter") != "false"

	response, err := s.db.QuerySignatures(query)
	if err != nil {
		fail(w, http.StatusInternalServerError, err, "failed to query signatures")
		return
	}

	s.filterResponse(response, shouldFilter)
	s.logSignatureResponse(r, response)

	succeed(w, response)
}

func (s *Service) serveImport(w http.ResponseWriter, r *http.Request) {
	var (
		req client.ImportRequest
		res client.ImportResponse
		err error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, err, "failed to decode body")
		return
	}

	res, err = s.importRaw(req)

	if err != nil {
		fail(w, http.StatusInternalServerError, err, "failed to import")
		return
	}

	s.logImportResponse(r, res)

	succeed(w, res)
}

func (s *Service) filterResponse(response client.SignatureResponse, shouldFilter bool) {
	s.canonicalSignaturesLock.RLock()

	for hash, values := range response[client.SignatureTypeFunction] {
		if expected, ok := s.canonicalSignatures[hash]; ok {
			for _, value := range values {
				value.Filtered = value.Name != expected
			}
		}
	}

	s.canonicalSignaturesLock.RUnlock()

	if shouldFilter {
		for hash, values := range response[client.SignatureTypeFunction] {
			var newValues []*client.SignatureData

			for _, value := range values {
				if !value.Filtered {
					newValues = append(newValues, value)
				}
			}

			response[client.SignatureTypeFunction][hash] = newValues
		}
	}
}

func (s *Service) logSignatureResponse(r *http.Request, response client.SignatureResponse) {
	fields := log.Fields{
		"ip": core.GetRemoteIP(r),
		"ua": core.GetUserAgent(r),
	}
	for typ, b := range response {
		var result []string
		for k, v := range b {
			var subresult []string
			for _, a := range v {
				subresult = append(subresult, a.Name)
			}
			result = append(result, fmt.Sprintf("%s=%s", k, strings.Join(subresult, ":")))
		}
		fields[string(typ)] = strings.Join(result, ";")
	}
	log.WithFields(fields).Infof("queried signatures")
}

func (s *Service) logImportResponse(r *http.Request, res client.ImportResponse) {
	fields := log.Fields{
		"ip": core.GetRemoteIP(r),
		"ua": core.GetUserAgent(r),
	}
	for _, typ := range client.SignatureTypes() {
		var imported []string
		var duplicated []string
		for k, v := range res[typ].Imported {
			imported = append(imported, fmt.Sprintf("%s=%s", k, v))
		}
		for k, v := range res[typ].Duplicated {
			duplicated = append(duplicated, fmt.Sprintf("%s=%s", k, v))
		}
		fields[fmt.Sprintf("%s_imported", typ)] = strings.Join(imported, ";")
		fields[fmt.Sprintf("%s_duplicated", typ)] = strings.Join(duplicated, ";")
		fields[fmt.Sprintf("%s_invalid", typ)] = strings.Join(res[typ].Invalid, ";")
	}
	log.WithFields(fields).Infof("imported signatures")
}

func (s *Service) serveStats(w http.ResponseWriter, r *http.Request) {
	resp := client.NewStatsResponse()

	var err error
	for _, typ := range client.SignatureTypes() {
		resp.Count[typ], err = s.db.CountSignatures(typ)
		if err != nil {
			fail(w, http.StatusInternalServerError, err, "failed to count signatures")
			return
		}
	}

	succeed(w, resp)
}

func (s *Service) serveRefreshCanonicalSignatures(w http.ResponseWriter, r *http.Request) {
	if err := s.loadCanonicalSignatures(); err != nil {
		fail(w, http.StatusInternalServerError, err, "failed to refresh")
		return
	}

	succeed(w, nil)
}

func (s *Service) serveExport(w http.ResponseWriter, r *http.Request) {
	s.dataExportLock.Lock()
	lastPath := s.dataExportPath
	s.dataExportLock.Unlock()

	if lastPath == "" {
		fail(w, http.StatusInternalServerError, nil, "export is not ready yet")
		return
	}

	f, err := os.Open(lastPath)
	if err != nil {
		fail(w, http.StatusInternalServerError, err, "failed to open file")
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		fail(w, http.StatusInternalServerError, err, "failed to stat file")
		return
	}

	fields := log.Fields{
		"ip": core.GetRemoteIP(r),
		"ua": core.GetUserAgent(r),
	}
	log.WithFields(fields).Infof("served export")

	w.Header().Set("Content-Disposition", `attachment; filename="export.txt"`)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, f); err != nil {
		fail(w, http.StatusInternalServerError, err, "failed to copy file")
		return
	}

	return

}

func (s *Service) startServer() {
	m := mux.NewRouter()
	m.HandleFunc("/v1/lookup", s.serveLookup).Methods("GET")
	m.HandleFunc("/v1/search", s.serveSearch).Methods("GET")
	m.HandleFunc("/v1/import", s.serveImport).Methods("POST")
	m.HandleFunc("/v1/stats", s.serveStats).Methods("GET")
	m.HandleFunc("/v1/export", s.serveExport).Methods("GET")
	m.HandleFunc("/v1/refresh_canonical_signatures", s.serveRefreshCanonicalSignatures).Methods("POST")

	cors := handlers.CORS(
		handlers.AllowedMethods([]string{"OPTIONS", "HEAD", "GET", "POST"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"}),
	)(m)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.HttpPort), cors); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Errorf("failed to listen and server")
		}
	}()
}
