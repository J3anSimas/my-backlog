package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"my-backlog/internal/apperrors"
)

func TestWriteError_UnknownError_Returns500WithGenericMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, errors.New("erro inesperado"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("esperado 500, obtido %d", rec.Code)
	}
	var body map[string]string
	if decErr := json.NewDecoder(rec.Body).Decode(&body); decErr != nil {
		t.Fatalf("body não é JSON válido: %v", decErr)
	}
	if body["message"] != "internal error" {
		t.Errorf("esperado mensagem genérica, obtido %q", body["message"])
	}
}

func TestWriteError_InfraError_Returns500WithGenericMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &apperrors.InfraError{Op: "db-insert", Cause: errors.New("connection refused")}

	writeError(rec, err)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("esperado 500, obtido %d", rec.Code)
	}
	var body map[string]string
	if decErr := json.NewDecoder(rec.Body).Decode(&body); decErr != nil {
		t.Fatalf("body não é JSON válido: %v", decErr)
	}
	if body["message"] != "internal error" {
		t.Errorf("esperado mensagem genérica, obtido %q", body["message"])
	}
	if body["message"] == "connection refused" {
		t.Error("detalhe de infra não deve vazar para o cliente")
	}
}

func TestWriteError_KindConflict_Returns409(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &apperrors.InputError{Kind: apperrors.KindConflict, Message: "e-mail já cadastrado"}

	writeError(rec, err)

	if rec.Code != http.StatusConflict {
		t.Errorf("esperado 409, obtido %d", rec.Code)
	}
}

func TestWriteError_KindNotFound_Returns404(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &apperrors.InputError{Kind: apperrors.KindNotFound, Message: "não encontrado"}

	writeError(rec, err)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obtido %d", rec.Code)
	}
}

func TestWriteError_KindUnauthorized_Returns401WithoutField(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &apperrors.InputError{Kind: apperrors.KindUnauthorized, Message: "credenciais inválidas"}

	writeError(rec, err)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, obtido %d", rec.Code)
	}
	var body map[string]string
	if decErr := json.NewDecoder(rec.Body).Decode(&body); decErr != nil {
		t.Fatalf("body não é JSON válido: %v", decErr)
	}
	if body["field"] != "" {
		t.Errorf("esperado field vazio, obtido %q", body["field"])
	}
}

func TestWriteError_KindInvalid_Returns422WithFieldAndMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &apperrors.InputError{Kind: apperrors.KindInvalid, Field: "email", Message: "email inválido"}

	writeError(rec, err)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("esperado 422, obtido %d", rec.Code)
	}
	var body map[string]string
	if decErr := json.NewDecoder(rec.Body).Decode(&body); decErr != nil {
		t.Fatalf("body não é JSON válido: %v", decErr)
	}
	if body["field"] != "email" {
		t.Errorf("esperado field=%q, obtido %q", "email", body["field"])
	}
	if body["message"] != "email inválido" {
		t.Errorf("esperado message=%q, obtido %q", "email inválido", body["message"])
	}
}
