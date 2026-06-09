package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"my-backlog/internal/apperrors"
)

var kindStatus = map[apperrors.InputKind]int{
	apperrors.KindInvalid:      http.StatusUnprocessableEntity,
	apperrors.KindNotFound:     http.StatusNotFound,
	apperrors.KindConflict:     http.StatusConflict,
	apperrors.KindUnauthorized: http.StatusUnauthorized,
}

type errorResponse struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// writeError mapeia erros de domínio para respostas HTTP JSON.
// InfraError é logado no servidor e retorna 500 sem vazar detalhes ao cliente.
func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var inputErr *apperrors.InputError
	if errors.As(err, &inputErr) {
		status, ok := kindStatus[inputErr.Kind]
		if !ok {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(errorResponse{Field: inputErr.Field, Message: inputErr.Message})
		return
	}

	var infraErr *apperrors.InfraError
	if errors.As(err, &infraErr) {
		log.Printf("infra error: %v", infraErr)
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errorResponse{Message: "internal error"})
}
