package app

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/smitendu1997/auto-message-dispatcher/config"
	"github.com/smitendu1997/auto-message-dispatcher/di"
	messagingGateway "github.com/smitendu1997/auto-message-dispatcher/gateway/messaging"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/smitendu1997/auto-message-dispatcher/utils"
	"github.com/smitendu1997/auto-message-dispatcher/utils/db"
	"github.com/smitendu1997/auto-message-dispatcher/utils/redis"
)

// App defines the core application interface
type App interface {
	// Core functionality
	GetRouter() *gin.Engine

	// Application name
	GetApplicationName() string

	// Database and connections
	GetConnections() *utils.Connections

	// Module management
	RegisterModule(module Module)
	GetModules() []Module

	// DI Container access
	GetContainer() *di.Container

	// Lifecycle methods
	Init(serviceName string)
	Bootstrap() error
	IsBootstrapped() bool

	// HTTP and Lambda handlers
	StartHTTPHandler(port string)
}

// BaseApp provides a standard implementation of the App interface
type BaseApp struct {
	Router          *gin.Engine
	ApplicationName string
	Connections     utils.Connections
	Container       *di.Container
	Modules         []Module
	Config          *config.AppConfig
	Bootstrapped    bool
}

// NewBaseApp creates a new BaseApp instance
func NewBaseApp(applicationName string, appConfig *config.AppConfig) *BaseApp {
	return &BaseApp{
		ApplicationName: applicationName,
		Router:          gin.Default(),
		Container:       di.NewContainer(),
		Config:          appConfig,
		Modules:         make([]Module, 0),
		Bootstrapped:    false,
	}
}

// GetRouter returns the gin router
func (a *BaseApp) GetRouter() *gin.Engine {
	return a.Router
}

// GetApplicationName returns the application name
func (a *BaseApp) GetApplicationName() string {
	return a.ApplicationName
}

// GetConnections returns the database and other connections
func (a *BaseApp) GetConnections() *utils.Connections {
	return &a.Connections
}

// GetContainer returns the DI container
func (a *BaseApp) GetContainer() *di.Container {
	return a.Container
}

// RegisterModule adds a module to the application
func (a *BaseApp) RegisterModule(module Module) {
	a.Modules = append(a.Modules, module)
}

// GetModules returns all registered modules
func (a *BaseApp) GetModules() []Module {
	return a.Modules
}

// IsBootstrapped returns whether the application has been bootstrapped
func (a *BaseApp) IsBootstrapped() bool {
	return a.Bootstrapped
}

func (a *BaseApp) Init(serviceName string) {

	const functionName = "main.App.Init"
	logger.Info(functionName, "initializing_app", a.ApplicationName)

	logger.SetLevel(-1)

	// Initialize database based on config
	mysqlDB, err := db.SetMySqlConnection(a.Config.Database.DSN)
	if err != nil {
		logger.Error(functionName, "failed to initialize MySQL connection:", err)
		os.Exit(1)
	}
	a.Connections.DB = mysqlDB

	// Initialize Redis connection
	redisClient, err := redis.SetRedisConnection(a.Config.Redis.URL)
	if err != nil {
		logger.Error(functionName, "failed to initialize Redis connection:", err)
		os.Exit(1)
	}
	a.Connections.Redis = redisClient

	// Add CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"*"}
	a.Router.Use(cors.New(config))

	// Register connections in the container
	a.Container.Register((*utils.Connections)(nil), &a.Connections)

	// Register all the http clients in the container
	a.registerHttpClients()

	logger.Info(functionName, "app_initialized")
}

func (a *BaseApp) registerHttpClients() {
	client := http.Client{}

	// Register http client for messaging in the container
	a.Container.Register((*messagingGateway.Messaging)(nil), &messagingGateway.Messaging{
		HttpClient: client,
		BaseUrl:    a.Config.MessagingApiUrl,
		ApiKey:     a.Config.MessagingApiKey,
	})

}

// Bootstrap configures all modules and sets up routes
func (a *BaseApp) Bootstrap() error {
	const functionName = "BaseApp.Bootstrap"

	if a.Bootstrapped {
		logger.Info(functionName, "already_bootstrapped")
		return nil
	}

	// Configure all modules
	for _, module := range a.Modules {
		logger.Info(functionName, "configuring_module", module.Name())
		module.Configure(a.Container)
	}

	// Register routes for all modules
	for _, module := range a.Modules {
		logger.Info(functionName, "registering_routes_module", module.Name())
		module.RegisterRoutes(a.Router, a.Container)
	}

	a.Bootstrapped = true
	logger.Info(functionName, "app_bootstrapped")
	return nil
}

// StartHTTPHandler starts the HTTP server
func (a *BaseApp) StartHTTPHandler(port string) {
	const functionName = "main.App.StartHTTPHandler"

	if !a.Bootstrapped {
		if err := a.Bootstrap(); err != nil {
			logger.Error(functionName, "failed to bootstrap app:", err)
			os.Exit(1)
		}
	}

	// Add health check endpoint
	a.Router.GET("/health", a.healthCheck)

	logger.Info(functionName, "starting_http_server", "port", port)
	err := http.ListenAndServe(":"+port, a.Router)
	if err != nil {
		logger.Error(functionName, err)
	}
}

// healthCheck provides a health endpoint
func (a *BaseApp) healthCheck(c *gin.Context) {
	const functionName = "main.App.healthCheck"

	health := map[string]string{
		"status": "up",
		"mysql":  "up",
		"redis":  "up",
	}

	// Check MySQL
	if err := a.Connections.DB.Ping(); err != nil {
		health["mysql"] = "down"
		health["status"] = "degraded"
		logger.Error(functionName, "MySQL health check failed:", err)
	}

	// Check Redis
	if err := a.Connections.Redis.Ping(c); err != nil {
		health["redis"] = "down"
		health["status"] = "degraded"
		logger.Error(functionName, "Redis health check failed:", err)
	}

	c.JSON(http.StatusOK, health)
}
