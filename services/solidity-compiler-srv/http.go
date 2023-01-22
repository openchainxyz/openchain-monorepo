package service

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/compiler"
	"github.com/openchainxyz/openchainxyz-monorepo/services/solidity-compiler-srv/client"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Service) serveCompile(w http.ResponseWriter, r *http.Request) {
	var request client.CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	solc, err := compiler.NewSolidityCompiler(request.Version)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	output, err := solc.CompileFromString(request.Code)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var abis []any
	for _, v := range output {
		abis = append(abis, v.Info.AbiDefinition)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&client.CompileResponse{
		Status: client.StatusSuccess,
		ABI:    abis,
	})
}

func (s *Service) startServer() {
	m := mux.NewRouter()
	m.HandleFunc("/v1/compile", s.serveCompile).Methods("POST")

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
