package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	handlerDto "github.com/smitendu1997/auto-message-dispatcher/handler/messaging/dto"
	"github.com/smitendu1997/auto-message-dispatcher/handler/messaging/poller"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	messagingService "github.com/smitendu1997/auto-message-dispatcher/services/messaging"
	"github.com/smitendu1997/auto-message-dispatcher/utils"
)

// MessageAPIHandler handles HTTP requests for Message API operations
type MessageAPIHandler interface {
	StartWorker() gin.HandlerFunc
	StopWorker() gin.HandlerFunc
	ListSentMessages() gin.HandlerFunc
}

type messageAPIHandler struct {
	handler          *poller.MessageHandler
	messagingService messagingService.MessagingSvcDriver
}

// NewMessageAPIHandler creates a new message API handler
func NewMessageAPIHandler(handler *poller.MessageHandler, messagingService messagingService.MessagingSvcDriver) MessageAPIHandler {
	return &messageAPIHandler{
		handler:          handler,
		messagingService: messagingService,
	}
}

// StartWorker starts the Message Poller
func (h *messageAPIHandler) StartWorker() gin.HandlerFunc {
	return func(c *gin.Context) {
		const functionName = "api.messageAPIHandler.StartWorker"

		err := h.handler.Start()
		if err != nil {
			logger.Error(functionName, "failed to start Message API:", err)
			response := utils.ResponseWithModel("500", "Failed to start Message API", nil)
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		logger.Info(functionName, "Message API started successfully")
		response := utils.ResponseWithModel("200", "Message API started successfully", nil)
		c.JSON(http.StatusOK, response)
	}
}

// StopWorker stops the Message Poller
func (h *messageAPIHandler) StopWorker() gin.HandlerFunc {
	return func(c *gin.Context) {
		const functionName = "api.messageAPIHandler.StopWorker"

		if !h.handler.IsRunning() {
			response := utils.ResponseWithModel("409", "Worker is not running", nil)
			c.JSON(http.StatusConflict, response)
			return
		}

		h.handler.Stop()
		logger.Info(functionName, "Message API stopped successfully")

		response := utils.ResponseWithModel("200", "Message API stopped successfully", map[string]interface{}{
			"status": "stopped",
		})
		c.JSON(http.StatusOK, response)
	}
}

// ListSentMessages lists all sent messages with pagination
func (h *messageAPIHandler) ListSentMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		const functionName = "api.messageAPIHandler.ListSentMessages"

		// Parse pagination parameters
		var paginationReq handlerDto.ListSentMessagesRequest
		if err := c.ShouldBindQuery(&paginationReq); err != nil {
			logger.Error(functionName, "failed to bind query parameters:", err)
			response := utils.ResponseWithModel("400", "Invalid query parameters", nil)
			c.JSON(http.StatusBadRequest, response)
			return
		}

		// Validate and set defaults
		paginationReq.ValidateAndSetDefaults()

		// Get sent messages from service
		messages, totalCount, hasMore, err := h.messagingService.ListSentMessages(c.Request.Context(), paginationReq.Limit, paginationReq.Offset)
		if err != nil {
			logger.Error(functionName, "failed to get sent messages:", err)
			response := utils.ResponseWithModel("500", "Failed to retrieve sent messages", nil)
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		// Convert domain messages to handler response format
		messageResponses := handlerDto.ConvertServiceResponseToHandlerResponse(messages)
		// Build response
		sentMessagesResponse := handlerDto.ListSentMessagesResponse{
			Messages: messageResponses,
			Pagination: handlerDto.PaginationMetadata{
				Limit:   paginationReq.Limit,
				Offset:  paginationReq.Offset,
				Total:   totalCount,
				HasMore: hasMore,
			},
		}

		logger.Info(functionName, "sent messages retrieved successfully", "count", len(messages), "total", totalCount)
		response := utils.ResponseWithModel("200", "Sent messages retrieved successfully", sentMessagesResponse)
		c.JSON(http.StatusOK, response)
	}
}
