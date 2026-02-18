package main

import (
	"encoding/json"
	"errors"
	"testing"
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
