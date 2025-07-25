package dto

import (
	"database/sql"

	"github.com/smitendu1997/auto-message-dispatcher/domain"
	"github.com/smitendu1997/auto-message-dispatcher/models/db"
)

func ConvertUpdateMessageParamsToDomain(r *domain.MessageDomain) db.UpdateMessageParams {
	return db.UpdateMessageParams{
		ID:        r.ID,
		Content:   sql.NullString{String: r.Content, Valid: r.Content != ""},
		Status:    sql.NullString{String: string(r.Status), Valid: r.Status.IsValid()},
		MessageId: sql.NullString{String: *r.MessageID, Valid: r.MessageID != nil},
		SentAt:    sql.NullTime{Time: *r.SentAt, Valid: r.SentAt != nil},
	}
}

// PaginationRequest represents pagination parameters for listing messages
type PaginationRequest struct {
	Limit  int32 `json:"limit" form:"limit"`
	Offset int32 `json:"offset" form:"offset"`
}

// ValidateAndSetDefaults validates pagination parameters and sets defaults
func (p *PaginationRequest) ValidateAndSetDefaults() {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 10 // default limit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}
