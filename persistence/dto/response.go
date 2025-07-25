package dto

import (
	"github.com/smitendu1997/auto-message-dispatcher/domain"
	"github.com/smitendu1997/auto-message-dispatcher/models/db"
)

func ConvertGetPendingMessageRowToMessageDomain(row *db.GetPendingMessageRow) *domain.MessageDomain {
	result := &domain.MessageDomain{
		ID:             row.ID,
		RecipientPhone: row.RecipientPhone,
		Content:        row.Content,
		Status:         domain.MessageStatus(row.Status),
	}
	if row.Messageid.Valid {
		result.MessageID = &row.Messageid.String
	}
	if row.SentAt.Valid {
		result.SentAt = &row.SentAt.Time
	}
	return result
}

// SentMessagesResponse represents the response for listing sent messages
type SentMessagesResponse struct {
	Messages   []*domain.MessageDomain `json:"messages"`
	Pagination PaginationMetadata      `json:"pagination"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	Limit   int32 `json:"limit"`
	Offset  int32 `json:"offset"`
	Total   int64 `json:"total"`
	HasMore bool  `json:"has_more"`
}

func ConvertGetSentMessageRowToMessageDomain(row *db.GetSentMessagesRow) *domain.MessageDomain {
	result := &domain.MessageDomain{
		ID:             row.ID,
		RecipientPhone: row.RecipientPhone,
		Content:        row.Content,
		Status:         domain.MessageStatus(row.Status),
	}
	if row.Messageid.Valid {
		result.MessageID = &row.Messageid.String
	}
	if row.SentAt.Valid {
		result.SentAt = &row.SentAt.Time
	}
	return result
}
