package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/chuxorg/chux-yanzi-emitter/internal/client"
)

var errInvalidInput = errors.New("invalid input")

// inputPayload is the stdin JSON schema for the emitter.
type inputPayload struct {
	Author     string          `json:"author"`
	SourceType string          `json:"source_type"`
	Title      *string         `json:"title,omitempty"`
	Prompt     string          `json:"prompt"`
	Response   string          `json:"response"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	PrevHash   *string         `json:"prev_hash,omitempty"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if len(input) == 0 {
		return fmt.Errorf("read stdin: %w", errInvalidInput)
	}

	var payload inputPayload
	if err := json.Unmarshal(input, &payload); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	if err := validate(payload); err != nil {
		return err
	}

	req := client.IntentRequest{
		Author:     payload.Author,
		SourceType: payload.SourceType,
		Title:      payload.Title,
		Prompt:     payload.Prompt,
		Response:   payload.Response,
		Meta:       payload.Meta,
		PrevHash:   payload.PrevHash,
	}

	resp, err := client.PostIntent(context.Background(), client.DefaultEndpoint, req)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n", resp.ID, resp.Hash)
	return nil
}

func validate(payload inputPayload) error {
	if payload.Author == "" {
		return fmt.Errorf("author: %w", errInvalidInput)
	}
	if payload.SourceType == "" {
		return fmt.Errorf("source_type: %w", errInvalidInput)
	}
	if payload.Prompt == "" {
		return fmt.Errorf("prompt: %w", errInvalidInput)
	}
	if payload.Response == "" {
		return fmt.Errorf("response: %w", errInvalidInput)
	}
	return nil
}
