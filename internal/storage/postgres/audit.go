package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

type auditEvent struct {
	WorkspaceID string
	ProjectID   string
	Action      string
	EntityType  string
	EntityID    string
	ActorUserID string
	ActorLabel  string
	BeforeState any
	AfterState  any
	Metadata    any
}

func (s *Store) recordEventTx(ctx context.Context, tx pgx.Tx, evt auditEvent) error {
	beforeState, err := marshalAuditJSON(evt.BeforeState)
	if err != nil {
		return fmt.Errorf("marshal audit before state: %w", err)
	}
	afterState, err := marshalAuditJSON(evt.AfterState)
	if err != nil {
		return fmt.Errorf("marshal audit after state: %w", err)
	}
	metadata, err := marshalAuditJSON(evt.Metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}
	if metadata == nil {
		metadata = json.RawMessage(`{}`)
	}

	err = s.q.WithTx(tx).InsertEvent(ctx, db.InsertEventParams{
		WorkspaceID: evt.WorkspaceID,
		ProjectID:   evt.ProjectID,
		Action:      evt.Action,
		EntityType:  evt.EntityType,
		EntityID:    evt.EntityID,
		ActorUserID: evt.ActorUserID,
		ActorLabel:  evt.ActorLabel,
		BeforeState: beforeState,
		AfterState:  afterState,
		Metadata:    metadata,
	})
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}

	return nil
}

func marshalAuditJSON(value any) (json.RawMessage, error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	return json.RawMessage(data), nil
}
