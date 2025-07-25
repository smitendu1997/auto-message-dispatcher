package app

import (
	"github.com/gin-gonic/gin"

	"github.com/smitendu1997/auto-message-dispatcher/di"
)

// Module defines a self-contained feature module
type Module interface {
	// Name returns the name of this module
	Name() string

	// Configure sets up this module's dependencies in the container
	Configure(container *di.Container)

	// RegisterRoutes sets up this module's routes
	RegisterRoutes(router *gin.Engine, container *di.Container)
}
