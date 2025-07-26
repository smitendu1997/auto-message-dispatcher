package domain

import (
	"time"
)

type MessageDomain struct {
	ID             int64
	RecipientPhone string
	Content        string
	Status         MessageStatus
	MessageID      *string
	SentAt         *time.Time
	RetryCount     int
}

func (m *MessageDomain) IsMessageSent() bool {
	return m.Status == MessageStatusSent
}
