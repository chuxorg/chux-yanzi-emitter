package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/chuxorg/chux-yanzi-emitter/internal/client"
)

func TestValidateCaptureArtifact(t *testing.T) {
	t.Parallel()

	payload := inputPayload{
		Author:     "alice",
		SourceType: "cli",
		Prompt:     "what is it",
		Response:   "a test",
	}

	if err := validate(payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateProjectArtifactAllowsEmptyPromptResponse(t *testing.T) {
	t.Parallel()

	meta := json.RawMessage(`{"artifact_type":"project"}`)
	payload := inputPayload{
		Author:     "alice",
		SourceType: "cli",
		Meta:       meta,
	}

	if err := validate(payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateMissingPromptResponseWithoutProjectMeta(t *testing.T) {
	t.Parallel()

	payload := inputPayload{
		Author:     "alice",
		SourceType: "cli",
	}

	err := validate(payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errInvalidInput) {
		t.Fatalf("expected errInvalidInput, got %v", err)
	}
}

func TestValidateRequiresAuthorEvenForProject(t *testing.T) {
	t.Parallel()

	meta := json.RawMessage(`{"artifact":{"type":"project"}}`)
	payload := inputPayload{
		SourceType: "cli",
		Meta:       meta,
	}

	err := validate(payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errInvalidInput) {
		t.Fatalf("expected errInvalidInput, got %v", err)
	}
}

func TestValidateRequiresSourceType(t *testing.T) {
	t.Parallel()

	payload := inputPayload{
		Author:   "alice",
		Prompt:   "hi",
		Response: "there",
	}

	err := validate(payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errInvalidInput) {
		t.Fatalf("expected errInvalidInput, got %v", err)
	}
}

func TestIsProjectArtifact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		meta string
		want bool
	}{
		{name: "empty", meta: "", want: false},
		{name: "invalid json", meta: "{", want: false},
		{name: "artifact_type", meta: `{"artifact_type":"project"}`, want: true},
		{name: "artifactType camel", meta: `{"artifactType":"project"}`, want: true},
		{name: "artifact object", meta: `{"artifact":{"type":"project"}}`, want: true},
		{name: "non project", meta: `{"artifact_type":"note"}`, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var meta json.RawMessage
			if tt.meta != "" {
				meta = json.RawMessage(tt.meta)
			}
			if got := isProjectArtifact(meta); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestRunWithDepsSuccess(t *testing.T) {
	input := `{"author":"alice","source_type":"cli","prompt":"hi","response":"there"}`

	var gotRequest client.IntentRequest
	postIntent := func(ctx context.Context, endpoint string, req client.IntentRequest) (client.IntentResponse, error) {
		if endpoint != "http://example" {
			t.Fatalf("expected endpoint http://example, got %s", endpoint)
		}
		gotRequest = req
		return client.IntentResponse{ID: "intent-1", Hash: "hash-1"}, nil
	}

	var output bytes.Buffer
	err := runWithDeps(strings.NewReader(input), &output, "http://example", postIntent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if output.String() != "intent-1 hash-1\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
	if gotRequest.Author != "alice" || gotRequest.SourceType != "cli" {
		t.Fatalf("unexpected request: %+v", gotRequest)
	}
}

func TestRunWithDepsEmptyInput(t *testing.T) {
	var output bytes.Buffer
	err := runWithDeps(strings.NewReader(""), &output, "http://example", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errInvalidInput) {
		t.Fatalf("expected errInvalidInput, got %v", err)
	}
	if !strings.Contains(err.Error(), "read stdin") {
		t.Fatalf("expected read stdin error, got %v", err)
	}
}

func TestRunWithDepsInvalidJSON(t *testing.T) {
	var output bytes.Buffer
	err := runWithDeps(strings.NewReader("{"), &output, "http://example", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "parse json") {
		t.Fatalf("expected parse json error, got %v", err)
	}
}

func TestRunWithDepsValidationError(t *testing.T) {
	var output bytes.Buffer
	err := runWithDeps(strings.NewReader(`{"source_type":"cli","prompt":"hi","response":"there"}`), &output, "http://example", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errInvalidInput) {
		t.Fatalf("expected errInvalidInput, got %v", err)
	}
}

func TestRunWithDepsPostIntentError(t *testing.T) {
	input := `{"author":"alice","source_type":"cli","prompt":"hi","response":"there"}`
	postIntent := func(ctx context.Context, endpoint string, req client.IntentRequest) (client.IntentResponse, error) {
		return client.IntentResponse{}, errors.New("post failed")
	}

	var output bytes.Buffer
	err := runWithDeps(strings.NewReader(input), &output, "http://example", postIntent)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "post failed") {
		t.Fatalf("expected post failed error, got %v", err)
	}
}

func TestRunWithDepsPreservesOptionalFields(t *testing.T) {
	input := `{"author":"alice","source_type":"cli","title":"greeting","prompt":"hi","response":"there","meta":{"artifact_type":"project","note":"hello"},"prev_hash":"prev-1"}`

	var gotRequest client.IntentRequest
	postIntent := func(ctx context.Context, endpoint string, req client.IntentRequest) (client.IntentResponse, error) {
		gotRequest = req
		return client.IntentResponse{ID: "intent-2", Hash: "hash-2"}, nil
	}

	var output bytes.Buffer
	err := runWithDeps(strings.NewReader(input), &output, "http://example", postIntent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gotRequest.Title == nil || *gotRequest.Title != "greeting" {
		t.Fatalf("expected title greeting, got %v", gotRequest.Title)
	}
	if gotRequest.PrevHash == nil || *gotRequest.PrevHash != "prev-1" {
		t.Fatalf("expected prev_hash prev-1, got %v", gotRequest.PrevHash)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(gotRequest.Meta, &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if meta["artifact_type"] != "project" || meta["note"] != "hello" {
		t.Fatalf("unexpected meta: %v", meta)
	}
}
