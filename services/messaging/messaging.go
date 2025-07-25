package messaging

import (
	"context"
	"encoding/json"

	"github.com/smitendu1997/auto-message-dispatcher/domain"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/spf13/cast"
)

type MessagingGateway interface {
	SendMessage(ctx context.Context, domainReq *domain.MessageDomain) (domain.MessageDomain, error)
}

func (c *MessagingSvc) PollAndProcessMessages(ctx context.Context) {
	const functionName = "messaging.MessagingSvc.pollAndProcessMessages"
	messages, err := c.MessagingPersistence.GetPendingMessages(ctx)
	if err != nil {
		logger.Error(functionName, "failed_to_get_pending_messages", err)
		return
	}

	for _, msg := range messages {
		logger.Info(functionName, "processing_message", msg.ID)
		response, err := c.Gateways.SendMessage(ctx, msg)
		if err != nil {
			logger.Error(functionName, "failed_to_send_message", err)
			continue
		}
		logger.Info(functionName, "message_sent_successfully")
		messageUpdateParams := domain.MessageDomain{
			ID:     msg.ID,
			Status: response.Status}
		msg.Status = response.Status
		logger.Info(functionName, "updating_message_status", msg.ID, msg.Status)
		// If the message was sent successfully, update the message in persistence
		// and cache the message ID in Redis for quick access
		if response.Status == domain.MessageStatusSent {
			logger.Info(functionName, "marking_message_as_sent")
			messageUpdateParams.MessageID = response.MessageID
			messageUpdateParams.SentAt = response.SentAt
			data := map[string]string{
				"response_message_id": *response.MessageID,
				"message_sent_at":     response.SentAt.Format("2006-01-02 15:04:05"),
			}
			jsonData, err := json.Marshal(data)
			c.Redis.Set(ctx, "messageSent_"+cast.ToString(msg.ID), jsonData, 0)
			if err != nil {
				logger.Error(functionName, "failed_to_cache_message_sent", err)
			}
		}

		err = c.MessagingPersistence.UpdateMessage(ctx, &messageUpdateParams)
		if err != nil {
			logger.Error(functionName, "failed_to_mark_message_as_sent", err)
		}
	}
}

func (c *MessagingSvc) ListSentMessages(ctx context.Context, limit, offset int32) ([]*domain.MessageDomain, int64, bool, error) {
	const functionName = "messaging.MessagingSvc.ListSentMessages"

	// Get sent messages from persistence layer
	messages, err := c.MessagingPersistence.GetSentMessages(ctx, limit, offset)
	if err != nil {
		logger.Error(functionName, "failed_to_get_sent_messages", err)
		return nil, 0, false, err
	}

	// Get total count for pagination metadata
	totalCount, err := c.MessagingPersistence.GetSentMessagesCount(ctx)
	if err != nil {
		logger.Error(functionName, "failed_to_get_sent_messages_count", err)
		return nil, 0, false, err
	}

	// Calculate hasMore
	hasMore := (offset + limit) < int32(totalCount)

	logger.Info(functionName, "sent_messages_retrieved_successfully", "count", len(messages), "total", totalCount, "hasMore", hasMore)

	return messages, totalCount, hasMore, nil
}
