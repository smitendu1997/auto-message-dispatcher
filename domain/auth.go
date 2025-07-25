package domain

type AuthenticationResponse struct {
	IsBasicAuthValidated bool
	Err                  error
}
