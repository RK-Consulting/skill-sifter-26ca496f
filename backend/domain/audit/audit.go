package audit

import (
	"database/sql"
	"encoding/json"
)

func Write(db *sql.DB, tenantID string, actorUserID int, entityType string, entityID int, action string, metadata interface{}) error {
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
