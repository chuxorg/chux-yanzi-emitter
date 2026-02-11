package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const DefaultEndpoint = "http://localhost:8080/v0/intents"

var ErrInvalidResponse = errors.New("invalid response")

// IntentRequest maps directly to the Yanzi Library POST /v0/intents API payload.
type IntentRequest struct {
	Author     string          `json:"author"`
	SourceType string          `json:"source_type"`
	Title      *string         `json:"title,omitempty"`
	Prompt     string          `json:"prompt"`
	Response   string          `json:"response"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	PrevHash   *string         `json:"prev_hash,omitempty"`
}

type IntentResponse struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}

func PostIntent(ctx context.Context, endpoint string, payload IntentRequest) (IntentResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return IntentResponse{}, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return IntentResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return IntentResponse{}, fmt.Errorf("post request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return IntentResponse{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		msg := string(bytes.TrimSpace(respBody))
		if msg == "" {
			msg = resp.Status
		}
		return IntentResponse{}, fmt.Errorf("request failed: %s", msg)
	}

	var out IntentResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return IntentResponse{}, fmt.Errorf("decode response: %w", err)
	}
	if out.ID == "" || out.Hash == "" {
		return IntentResponse{}, ErrInvalidResponse
	}

	return out, nil
}
