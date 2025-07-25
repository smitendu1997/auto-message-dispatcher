package dto

import (
	"time"

	"github.com/smitendu1997/auto-message-dispatcher/domain"
)

type MessagingResponse struct {
	Message   string `json:"message"`
	MessageId string `json:"messageId"`
}

type MessagingRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

var messageStatusMap = map[string]domain.MessageStatus{
	"Accepted": domain.MessageStatusSent,
	"Failed":   domain.MessageStatusFailed,
}

func ConvertMessageDomainToMessagingRequest(domainReq *domain.MessageDomain) *MessagingRequest {
	return &MessagingRequest{
		To:      domainReq.RecipientPhone,
		Content: domainReq.Content,
	}
}

func ConvertMessagingResponseToDomainResponse(gatewayResp *MessagingResponse) domain.MessageDomain {
	domainMessage := domain.MessageDomain{
		Status:    messageStatusMap[gatewayResp.Message], // Assuming status is sent for successful response
		MessageID: &gatewayResp.MessageId,
	}
	if domainMessage.IsMessageSent() {
		sentAt := time.Now()
		domainMessage.SentAt = &sentAt
	} else {
		domainMessage.SentAt = nil
	}

	return domainMessage
}
