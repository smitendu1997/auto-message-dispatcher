// Messaging represents a gateway for sending SMS and Push Notifications.
// It holds the configuration needed to communicate with the external service.
//
// Fields:
//   - ApiKey: Authentication key for the Messaging service
//   - BaseUrl: Base URL of the Messaging service API
//
// The gateway provides two main functionalities:
//   - SendSMS: Sends SMS messages using the configured service
//   - SendPN: Sends Push Notifications using the configured service
//
// Usage:
//
//	comm := NewCommunicationGateway()
//	err := comm.SendSMS(smsRequest)
//	err := comm.SendPN(pnRequest)
//
// The gateway automatically reads configuration from environment variables:
//   - COMMUNICATION_API_BASE_URL: Base URL for the Messaging service
//   - COMMUNICATION_API_KEY: API key for authentication
package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/smitendu1997/auto-message-dispatcher/domain"
	"github.com/smitendu1997/auto-message-dispatcher/gateway/dto"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	MessagingSvc "github.com/smitendu1997/auto-message-dispatcher/services/messaging"
	"github.com/smitendu1997/auto-message-dispatcher/utils"
)

type Messaging struct {
	ApiKey     string
	BaseUrl    string
	HttpClient http.Client
}

func NewMessagingGateway(httpClient http.Client, baseUrl, apiKey string) MessagingSvc.MessagingGateway {
	return &Messaging{
		BaseUrl:    baseUrl,
		ApiKey:     apiKey,
		HttpClient: httpClient,
	}
}

func (c *Messaging) SendMessage(ctx context.Context, domainReq *domain.MessageDomain) (domain.MessageDomain, error) {
	gatewayReq := dto.ConvertMessageDomainToMessagingRequest(domainReq)
	gatewayReqByte, err := json.Marshal(gatewayReq)
	if err != nil {
		return domain.MessageDomain{}, err
	}
	res, statusCode, err := utils.HttpCall("core.Messaging.SendSMS", ctx, "POST", c.BaseUrl, c.HttpClient, gatewayReqByte, map[string]string{
		"x-ins-auth-key": c.ApiKey,
		"Content-Type":   "application/json",
	})
	if err != nil {
		return domain.MessageDomain{}, err
	}
	var result dto.MessagingResponse
	if err := json.Unmarshal(res, &result); err != nil {
		return domain.MessageDomain{}, err
	}

	logger.Info("core.Messaging.SendSMS", "Req & Res: ", string(gatewayReqByte), string(res))
	if statusCode != 202 {
		return domain.MessageDomain{}, errors.New("failed to send SMS")
	}
	return dto.ConvertMessagingResponseToDomainResponse(&result), nil
}
