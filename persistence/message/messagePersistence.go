package message

import (
	"context"
	"database/sql"

	"github.com/smitendu1997/auto-message-dispatcher/dbdebug"
	"github.com/smitendu1997/auto-message-dispatcher/domain"
	"github.com/smitendu1997/auto-message-dispatcher/models/db"
	"github.com/smitendu1997/auto-message-dispatcher/persistence/dto"
)

func NewMessagePersistence(DB *sql.DB) MessagePersistence {
	querier := db.New(dbdebug.Wrap(DB))
	return &Message{Querier: querier}
}

type MessagePersistence interface {
	GetPendingMessages(ctx context.Context) ([]*domain.MessageDomain, error)
	UpdateMessage(ctx context.Context, msg *domain.MessageDomain) error
	GetSentMessages(ctx context.Context, limit, offset int32) ([]*domain.MessageDomain, error)
	GetSentMessagesCount(ctx context.Context) (int64, error)
}

type Message struct {
	Querier *db.Queries
}

func (m *Message) GetPendingMessages(ctx context.Context) ([]*domain.MessageDomain, error) {
	rows, err := m.Querier.GetPendingMessage(ctx)
	var messages []*domain.MessageDomain
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		message := dto.ConvertGetPendingMessageRowToMessageDomain(&row)
		messages = append(messages, message)
	}
	return messages, nil
}

func (m *Message) UpdateMessage(ctx context.Context, msg *domain.MessageDomain) error {
	err := m.Querier.UpdateMessage(ctx, dto.ConvertUpdateMessageParamsToDomain(msg))
	if err != nil {
		return err
	}
	return nil
}

func (m *Message) GetSentMessages(ctx context.Context, limit, offset int32) ([]*domain.MessageDomain, error) {
	rows, err := m.Querier.GetSentMessages(ctx, db.GetSentMessagesParams{
		Limit:  limit,
		Offset: offset,
	})
	var messages []*domain.MessageDomain
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		message := dto.ConvertGetSentMessageRowToMessageDomain(&row)
		messages = append(messages, message)
	}
	return messages, nil
}

func (m *Message) GetSentMessagesCount(ctx context.Context) (int64, error) {
	count, err := m.Querier.GetSentMessagesCount(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}
