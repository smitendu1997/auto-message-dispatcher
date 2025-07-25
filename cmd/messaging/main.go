package main

import (
	"github.com/spf13/viper"

	"github.com/smitendu1997/auto-message-dispatcher/app"
	"github.com/smitendu1997/auto-message-dispatcher/config"
	"github.com/smitendu1997/auto-message-dispatcher/handler/messaging"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
)

func main() {
	const functionName = "main.main"
	const serviceName = "go-message-scheduler"

	// Load application configuration
	appConfig := config.LoadConfig()

	// Create application
	application := app.NewBaseApp(serviceName, appConfig)

	// Initialize core components
	application.Init(serviceName)

	// Register modules based on configuration
	registerModules(application, appConfig)

	// Start HTTP handler
	application.StartHTTPHandler(viper.GetString("MESSAGING_API_PORT"))
	logger.Info(functionName, "app_started")
}

func registerModules(app app.App, appConfig *config.AppConfig) {
	// Register SQS worker module
	app.RegisterModule(messaging.NewModule())
}
