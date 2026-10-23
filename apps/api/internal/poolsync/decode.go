package poolsync

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type body struct {
	FacileID        string  `json:"facile_id"`
	ProjectFacileID string  `json:"project_facile_id"`
	Name            string  `json:"name"`
	Title           string  `json:"title"`
	Description     *string `json:"description"`
	Icon            *string `json:"icon"`
	Status          string  `json:"status"`
	ActorEmail      string  `json:"actor_email"`
}

type wrapped struct {
	App            string `json:"app"`
	FacileID       string `json:"facile_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Timestamp      string `json:"timestamp"`
	Payload        *body  `json:"payload"`
}

type inbound struct {
	body
	app      string
	key      string
	facileID string
	at       time.Time
}

func decode(raw json.RawMessage) (inbound, error) {
	var env wrapped
	if err := json.Unmarshal(raw, &env); err != nil {
		return inbound{}, fmt.Errorf("decode pool event: %w", err)
	}
	evt := inbound{app: env.App, key: env.IdempotencyKey}
	if at, err := time.Parse(time.RFC3339, env.Timestamp); err == nil {
		evt.at = at
	}
	if env.Payload != nil {
		evt.body = *env.Payload
	} else if err := json.Unmarshal(raw, &evt.body); err != nil {
		return inbound{}, fmt.Errorf("decode bare pool payload: %w", err)
	}
	evt.facileID = strings.TrimSpace(evt.body.FacileID)
	if evt.facileID == "" {
		evt.facileID = strings.TrimSpace(env.FacileID)
	}
	if strings.TrimSpace(evt.Name) == "" {
		evt.Name = evt.Title
	}
	evt.Name = strings.TrimSpace(evt.Name)
	evt.Status = strings.ToLower(strings.TrimSpace(evt.Status))
	return evt, nil
}

func isLegacyID(id string) bool {
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return id != ""
}
