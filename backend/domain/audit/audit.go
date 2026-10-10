package audit

import (
	"database/sql"
	"encoding/json"
)

type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func Write(db *sql.DB, tenantID string, actorUserID int, entityType string, entityID int, action string, metadata interface{}) error {
	return write(db, tenantID, actorUserID, entityType, entityID, action, metadata)
}

func WriteTx(tx *sql.Tx, tenantID string, actorUserID int, entityType string, entityID int, action string, metadata interface{}) error {
	return write(tx, tenantID, actorUserID, entityType, entityID, action, metadata)
}

func write(execer execer, tenantID string, actorUserID int, entityType string, entityID int, action string, metadata interface{}) error {
	var payload []byte
	var err error
	if metadata == nil {
		payload = []byte("{}")
	} else {
		payload, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}
	var actor interface{}
	if actorUserID > 0 {
		actor = actorUserID
	}
	_, err = db.Exec(`
		INSERT INTO audit_events (
			tenant_id, actor_user_id, entity_type, entity_id, action, metadata
		) VALUES ($1,$2,$3,$4,$5,$6::jsonb)
	`, tenantID, actor, entityType, entityID, action, string(payload))
	return err
}
