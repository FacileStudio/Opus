package poolsync

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type spaceBody struct {
	RegistreSpaceID string `json:"registre_space_id"`
	Personal        bool   `json:"personal"`
	OwnerSubject    string `json:"owner_subject"`
	OwnerEmail      string `json:"owner_email"`
}

type taskBody struct {
	ProjectFacileID string          `json:"project_facile_id"`
	Status          string          `json:"status"`
	Priority        *string         `json:"priority"`
	DueDate         json.RawMessage `json:"due_date"`
	ActorEmail      string          `json:"actor_email"`
	ActorSubject    string          `json:"actor_subject"`
}

type timeBody struct {
	TaskFacileID     string  `json:"task_facile_id"`
	UserEmail        string  `json:"user_email"`
	UserSubject      string  `json:"user_subject"`
	StartedAt        string  `json:"started_at"`
	StoppedAt        *string `json:"stopped_at"`
	PausedDurationMS int64   `json:"paused_duration_ms"`
}

type body struct {
	FacileID    string     `json:"facile_id"`
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Icon        *string    `json:"icon"`
	Space       *spaceBody `json:"space"`
	UpdatedAt   string     `json:"updated_at"`
	taskBody
	timeBody
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
	updated  time.Time
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
	evt.updated = evt.at
	if updated, err := time.Parse(time.RFC3339Nano, evt.UpdatedAt); err == nil {
		evt.updated = updated
	}
	return evt, nil
}

// dueDate reports the due date the payload sets, nil to clear it, and false
// when the payload carries none, which leaves the stored one alone.
func (b taskBody) dueDate() (*time.Time, bool) {
	raw := strings.TrimSpace(string(b.DueDate))
	if raw == "" {
		return nil, false
	}
	if raw == "null" {
		return nil, true
	}
	var text string
	if err := json.Unmarshal(b.DueDate, &text); err != nil {
		return nil, false
	}
	if text = strings.TrimSpace(text); text == "" {
		return nil, true
	}
	for _, layout := range []string{time.DateOnly, time.RFC3339} {
		if day, err := time.Parse(layout, text); err == nil {
			return &day, true
		}
	}
	return nil, false
}

func isLegacyID(id string) bool {
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return id != ""
}
