package messaging

import (
	"context"
	"encoding/json"
	"time"

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
		// Check Redis cache first
		cacheKey := "messageSent_" + cast.ToString(msg.ID)
		if sentData, err := c.checkMessageCache(ctx, cacheKey); err == nil && sentData != nil {
			c.handleAlreadySentMessage(ctx, functionName, *msg, sentData)
			continue
		}

		// Send message through gateway
		response, err := c.Gateways.SendMessage(ctx, msg)
		if err != nil {
			c.handleFailedMessage(ctx, functionName, *msg)
			return
		}

		// Update message status based on response
		c.updateMessageStatus(ctx, functionName, *msg, &response)
	}
}

func (c *MessagingSvc) checkMessageCache(ctx context.Context, cacheKey string) (map[string]string, error) {
	cached, err := c.Redis.Get(ctx, cacheKey)
	if err != nil || len(cached) == 0 {
		return nil, err
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(cached), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *MessagingSvc) handleAlreadySentMessage(ctx context.Context, functionName string, msg domain.MessageDomain, sentData map[string]string) {
	logger.Info(functionName, "message_already_sent", msg.ID)

	updateParams := domain.MessageDomain{
		ID:     msg.ID,
		Status: domain.MessageStatusSent,
	}

	if messageID := sentData["response_message_id"]; messageID != "" {
		updateParams.MessageID = &messageID
	}

	if sentAt := sentData["message_sent_at"]; sentAt != "" {
		if parsedTime, err := time.Parse(time.DateTime, sentAt); err == nil {
			updateParams.SentAt = &parsedTime
		}
	}

	if err := c.MessagingPersistence.UpdateMessage(ctx, &updateParams); err != nil {
		logger.Error(functionName, "failed_to_update_message_status", err)
	}
}

func (c *MessagingSvc) handleFailedMessage(ctx context.Context, functionName string, msg domain.MessageDomain) {
	logger.Error(functionName, "failed_to_send_message", msg.ID)

	updateParams := domain.MessageDomain{
		ID:     msg.ID,
		Status: domain.MessageStatusFailed,
	}

	if err := c.MessagingPersistence.UpdateMessage(ctx, &updateParams); err != nil {
		logger.Error(functionName, "failed_to_update_failed_message_status", err)
	}
}

func (c *MessagingSvc) updateMessageStatus(ctx context.Context, functionName string, msg domain.MessageDomain, response *domain.MessageDomain) {
	logger.Info(functionName, "message_sent_successfully", msg.ID)

	updateParams := domain.MessageDomain{
		ID:     msg.ID,
		Status: response.Status,
	}

	if response.Status == domain.MessageStatusSent {
		updateParams.MessageID = response.MessageID
		updateParams.SentAt = response.SentAt

		if err := c.cacheSentMessage(ctx, msg.ID, response); err != nil {
			logger.Error(functionName, "failed_to_cache_message_sent", err)
		}
	}

	if err := c.MessagingPersistence.UpdateMessage(ctx, &updateParams); err != nil {
		logger.Error(functionName, "failed_to_update_message_status", err)
	}
}

func (c *MessagingSvc) cacheSentMessage(ctx context.Context, msgID int64, response *domain.MessageDomain) error {
	data := map[string]string{
		"response_message_id": *response.MessageID,
		"message_sent_at":     response.SentAt.Format(time.DateTime),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.Redis.Set(ctx, "messageSent_"+cast.ToString(msgID), jsonData, 24*time.Hour)
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
