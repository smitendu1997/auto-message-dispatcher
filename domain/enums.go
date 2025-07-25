package domain

// MessageStatus represents the status of a message in the system.
type MessageStatus string

const (
	MessageStatusPending MessageStatus = "pending"
	MessageStatusSent    MessageStatus = "sent"
	MessageStatusFailed  MessageStatus = "failed"
)

// IsValid checks if the value is a valid MessageStatus
func (s MessageStatus) IsValid() bool {
	switch s {
	case MessageStatusPending, MessageStatusSent, MessageStatusFailed:
		return true
	}
	return false
}
