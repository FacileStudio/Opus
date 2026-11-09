package github

import "testing"

func TestDecodePayloadBare(t *testing.T) {
	raw := []byte(`{"action":"opened","pull_request":{"id":1,"number":42,"title":"test"}}`)
	payload, err := DecodePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.PullRequest == nil || payload.PullRequest.Number != 42 {
		t.Fatalf("expected PR 42, got %+v", payload.PullRequest)
	}
}

func TestDecodePayloadWrapped(t *testing.T) {
	raw := []byte(`{"app":"antenne","channel":"git.pull_request","payload":{"action":"opened","pull_request":{"id":1,"number":99,"title":"wrapped test"}}}`)
	payload, err := DecodePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.PullRequest == nil || payload.PullRequest.Number != 99 {
		t.Fatalf("expected PR 99, got %+v", payload.PullRequest)
	}
}
