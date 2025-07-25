package dto

import (
	"time"

	"github.com/smitendu1997/auto-message-dispatcher/domain"
)

// Handler DTOs for List Sent Messages API
type ListSentMessagesRequest struct {
	Limit  int32 `json:"limit" form:"limit"`
	Offset int32 `json:"offset" form:"offset"`
}

// ValidateAndSetDefaults validates pagination parameters and sets defaults
func (r *ListSentMessagesRequest) ValidateAndSetDefaults() {
	if r.Limit <= 0 {
		r.Limit = -1 // default limit
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
}

type MessageResponse struct {
	ID             int64      `json:"id"`
	RecipientPhone string     `json:"recipient_phone"`
	Content        string     `json:"content"`
	Status         string     `json:"status"`
	MessageID      *string    `json:"message_id,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	CreatedOn      *time.Time `json:"created_on,omitempty"`
}

type PaginationMetadata struct {
	Limit   int32 `json:"limit"`
	Offset  int32 `json:"offset"`
	Total   int64 `json:"total"`
	HasMore bool  `json:"has_more"`
}

type ListSentMessagesResponse struct {
	Messages   []MessageResponse  `json:"messages"`
	Pagination PaginationMetadata `json:"pagination"`
}

// ConvertDomainToMessageResponse converts domain message to handler response
func ConvertDomainToMessageResponse(msg *domain.MessageDomain) MessageResponse {
	return MessageResponse{
		ID:             msg.ID,
		RecipientPhone: msg.RecipientPhone,
		Content:        msg.Content,
		Status:         string(msg.Status),
		MessageID:      msg.MessageID,
		SentAt:         msg.SentAt,
	}
}

// ConvertServiceResponseToHandlerResponse converts service response to handler response
func ConvertServiceResponseToHandlerResponse(serviceResp []*domain.MessageDomain) []MessageResponse {
	var messages []MessageResponse
	for _, msg := range serviceResp {
		messages = append(messages, ConvertDomainToMessageResponse(msg))
	}
	return messages
}
