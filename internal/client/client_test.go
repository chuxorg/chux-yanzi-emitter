package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostIntentSuccess(t *testing.T) {
	var gotPayload IntentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", ct)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &gotPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"intent-1","hash":"hash-1"}`))
	}))
	defer server.Close()

	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = originalClient })

	payload := IntentRequest{
		Author:     "alice",
		SourceType: "cli",
		Prompt:     "hi",
		Response:   "there",
	}

	resp, err := PostIntent(context.Background(), server.URL, payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ID != "intent-1" || resp.Hash != "hash-1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if gotPayload.Author != payload.Author || gotPayload.SourceType != payload.SourceType {
		t.Fatalf("unexpected payload: %+v", gotPayload)
	}
}

func TestPostIntentNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad input", http.StatusBadRequest)
	}))
	defer server.Close()

	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = originalClient })

	_, err := PostIntent(context.Background(), server.URL, IntentRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "bad input") {
		t.Fatalf("expected error to contain response body, got %v", err)
	}
}

func TestPostIntentInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":`))
	}))
	defer server.Close()

	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = originalClient })

	_, err := PostIntent(context.Background(), server.URL, IntentRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("expected decode response error, got %v", err)
	}
}

func TestPostIntentMissingFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"","hash":""}`))
	}))
	defer server.Close()

	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = originalClient })

	_, err := PostIntent(context.Background(), server.URL, IntentRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrInvalidResponse {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}
