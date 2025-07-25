package core

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
	"github.com/smitendu1997/auto-message-dispatcher/domain"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/smitendu1997/auto-message-dispatcher/utils"
	"github.com/spf13/viper"
)

type Authentication interface {
	Authorize(ctx context.Context, AuthToken string) (*domain.AuthenticationResponse, error)
}

type AuthenticationSvc struct {
}

type AuthenticationDetails struct {
	Ctx       context.Context
	AuthToken string
}

// AuthenticationSVC creates a new instance of Authentication service
func AuthenticationSVC() Authentication {
	return &AuthenticationSvc{}
}

// Authorize ...
func (svc *AuthenticationSvc) Authorize(ctx context.Context, AuthToken string) (*domain.AuthenticationResponse, error) {
	ad := &AuthenticationDetails{
		Ctx:       ctx,
		AuthToken: AuthToken,
	}

	authresp := ad.validateBasicAuthWithIPCheck()
	return &authresp, authresp.Err
}

// ValidateBasicAuthWithIPCheck validates basic auth credentials and checks if client IP is in whitelist
func (ad *AuthenticationDetails) validateBasicAuthWithIPCheck() domain.AuthenticationResponse {
	const functionName = "services.core.AuthenticationSvc.ValidateBasicAuthWithIPCheck"

	// Extract and validate basic auth credentials
	authHeader := ad.AuthToken
	if !strings.HasPrefix(authHeader, "Basic ") {
		logger.Error(functionName, "Invalid auth header format")
		return domain.AuthenticationResponse{
			Err: errors.New("Invalid authentication header"),
		}
	}

	// Remove "Basic " prefix and decode
	credentials, err := base64.StdEncoding.DecodeString(authHeader[6:])
	if err != nil {
		logger.Error(functionName, "Failed to decode auth header: ", err)
		return domain.AuthenticationResponse{
			Err: errors.New("Invalid authentication credentials"),
		}
	}

	// Split username and password
	parts := strings.SplitN(string(credentials), ":", 2)
	if len(parts) != 2 {
		logger.Error(functionName, "Invalid credentials format")
		return domain.AuthenticationResponse{
			Err: errors.New("Invalid credentials format"),
		}
	}

	username := parts[0]
	password := parts[1]

	// Check both current and previous credentials to support rotation
	if isValidCredential(username, password) {
		logger.Info(functionName, "API authentication successful for: ", username)
		return domain.AuthenticationResponse{
			IsBasicAuthValidated: true,
		}
	}

	logger.Error(functionName, "Invalid credentials for: ", username)
	return domain.AuthenticationResponse{
		Err: errors.New("Invalid username or password"),
	}
}

// isValidCredential checks if the username and password are valid
// This supports credential rotation by checking both current and previous credentials
func isValidCredential(username, password string) bool {
	// Get username/password pairs from environment variables
	auth := utils.SHA256Hash(username + ":" + password)

	// Current credentials format: API_USER_CURRENT=username:password
	currentCredentialsEnv := viper.GetString("API_USER_CURRENT")
	currentCredentialsEnvList := strings.Split(currentCredentialsEnv, ",")

	// Previous credentials format: API_USER_PREVIOUS=username:password
	previousCredentialsEnv := viper.GetString("API_USER_PREVIOUS")
	previousCredentialsEnvList := strings.Split(previousCredentialsEnv, ",")

	// Parse the current credentials
	if len(currentCredentialsEnvList) > 0 && utils.InArraySlice(currentCredentialsEnvList, auth) {
		return true
	}

	// Parse the previous credentials (for rotation support)
	if len(previousCredentialsEnvList) > 0 && utils.InArraySlice(previousCredentialsEnvList, auth) {
		return true
	}

	return false
}
