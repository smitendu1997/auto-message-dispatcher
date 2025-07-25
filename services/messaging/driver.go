package messaging

import (
	"context"

	"github.com/smitendu1997/auto-message-dispatcher/domain"
	"github.com/smitendu1997/auto-message-dispatcher/persistence/message"
	"github.com/smitendu1997/auto-message-dispatcher/utils/redis"
)

func NewMessagingSvc(Gateways MessagingGateway, MessagingPersistence message.MessagePersistence, Redis *redis.RedisClient) MessagingSvcDriver {

	// Persistence Declarations
	return &MessagingSvc{
		MessagingPersistence: MessagingPersistence,
		Gateways:             Gateways,
		Redis:                Redis,
	}

}

type MessagingSvcDriver interface {
	PollAndProcessMessages(ctx context.Context)
	ListSentMessages(ctx context.Context, limit, offset int32) ([]*domain.MessageDomain, int64, bool, error)
}

type MessagingSvc struct {
	MessagingPersistence message.MessagePersistence
	Gateways             MessagingGateway
	Redis                *redis.RedisClient
}
