package messaging

import (
	"github.com/gin-gonic/gin"
	"github.com/smitendu1997/auto-message-dispatcher/di"
	messagingGateway "github.com/smitendu1997/auto-message-dispatcher/gateway/messaging"
	"github.com/smitendu1997/auto-message-dispatcher/handler/messaging/api"
	"github.com/smitendu1997/auto-message-dispatcher/handler/messaging/poller"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/smitendu1997/auto-message-dispatcher/middleware"
	messagePersistence "github.com/smitendu1997/auto-message-dispatcher/persistence/message"
	coreService "github.com/smitendu1997/auto-message-dispatcher/services/core"
	messagingService "github.com/smitendu1997/auto-message-dispatcher/services/messaging"
	"github.com/smitendu1997/auto-message-dispatcher/utils"
)

// Module implements the Message poller functionality
type Module struct {
}

// NewModule creates a new Message poller module
func NewModule() *Module {
	return &Module{}
}

// Name returns the module name
func (m *Module) Name() string {
	return "messaging"
}

// Configure sets up dependencies for this module
func (m *Module) Configure(container *di.Container) {
	const functionName = "messaging.Module.Configure"
	logger.Info(functionName, "configuring_messaging_module")

	// Register gateway implementations
	container.RegisterFactory((*messagingService.MessagingGateway)(nil), func(c *di.Container) interface{} {
		messagingGatewayConfig := c.Resolve((*messagingGateway.Messaging)(nil)).(*messagingGateway.Messaging)
		return messagingGateway.NewMessagingGateway(messagingGatewayConfig.HttpClient, messagingGatewayConfig.BaseUrl, messagingGatewayConfig.ApiKey)
	})

	// Register persistence layers
	container.RegisterFactory((*messagePersistence.MessagePersistence)(nil), func(c *di.Container) interface{} {
		connections := c.Resolve((*utils.Connections)(nil)).(*utils.Connections)
		return messagePersistence.NewMessagePersistence(connections.DB.DB())
	})

	// Register Authentication service
	container.RegisterFactory((*coreService.Authentication)(nil), func(c *di.Container) interface{} {
		return coreService.AuthenticationSVC()
	})

	// Register Message service
	container.RegisterFactory((*messagingService.MessagingSvcDriver)(nil), func(c *di.Container) interface{} {
		messagingGateway := c.Resolve((*messagingService.MessagingGateway)(nil)).(messagingService.MessagingGateway)
		messagingRepo := c.Resolve((*messagePersistence.MessagePersistence)(nil)).(messagePersistence.MessagePersistence)
		connections := c.Resolve((*utils.Connections)(nil)).(*utils.Connections)
		return messagingService.NewMessagingSvc(messagingGateway, messagingRepo, connections.Redis)
	})

	// Register Message poller
	container.RegisterFactory((*poller.MessageHandler)(nil), func(c *di.Container) interface{} {
		messagingService := c.Resolve((*messagingService.MessagingSvcDriver)(nil)).(messagingService.MessagingSvcDriver)
		return poller.NewMessageHandler(messagingService)
	})

	// Register API handler factory
	container.RegisterFactory((*api.MessageAPIHandler)(nil), func(c *di.Container) interface{} {
		messageHandler := c.Resolve((*poller.MessageHandler)(nil)).(*poller.MessageHandler)
		messagingService := c.Resolve((*messagingService.MessagingSvcDriver)(nil)).(messagingService.MessagingSvcDriver)
		return api.NewMessageAPIHandler(messageHandler, messagingService)
	})

	logger.Info(functionName, "message_api_module_configured")
}

// RegisterRoutes sets up routes for this module
func (m *Module) RegisterRoutes(router *gin.Engine, container *di.Container) {
	const functionName = "messageAPI.Module.RegisterRoutes"
	logger.Info(functionName, "registering_routes")

	// Resolve the API handler
	handler := container.Resolve((*api.MessageAPIHandler)(nil)).(api.MessageAPIHandler)
	authSvc := container.Resolve((*coreService.Authentication)(nil)).(coreService.Authentication)

	// Setup API routes
	apiGroup := router.Group("/messaging")
	apiGroup.Use(middleware.AuthorizationMiddleware(authSvc))
	apiGroup.Use(middleware.APICallLogsMiddleware)

	actionGroup := apiGroup.Group("/action")
	{
		actionGroup.POST("/start", handler.StartWorker())
		actionGroup.POST("/stop", handler.StopWorker())
	}

	listGroup := apiGroup.Group("/list")
	{
		listGroup.GET("/sent", handler.ListSentMessages())
	}

	// Auto-start the Message poller after routes are registered (all dependencies are ready)
	go func() {
		logger.Info(functionName, "auto_starting_message_handler")
		messagePoller := container.Resolve((*poller.MessageHandler)(nil)).(*poller.MessageHandler)
		if messagePoller != nil {
			err := messagePoller.Start()
			if err != nil {
				logger.Error(functionName, "failed_to_auto_start_message_poller", err)
			} else {
				logger.Info(functionName, "message_poller_auto_started_successfully")
			}
		} else {
			logger.Error(functionName, "message_poller_instance_not_found")
		}
	}()

	logger.Info(functionName, "routes_registered")
}
