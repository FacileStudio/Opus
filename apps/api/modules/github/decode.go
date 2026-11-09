package github

import "encoding/json"

// DecodePayload decodes a webhook payload which may be bare or wrapped in an envelope.
func DecodePayload(raw []byte) (*WebhookPayload, error) {
	var wrapped struct {
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Payload) > 0 && string(wrapped.Payload) != "null" {
		var payload WebhookPayload
		if err := json.Unmarshal(wrapped.Payload, &payload); err == nil && isUsablePayload(&payload) {
			return &payload, nil
		}
	}

	var payload WebhookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func isUsablePayload(p *WebhookPayload) bool {
	return p.PullRequest != nil || p.Ref != "" || len(p.Commits) > 0
}
