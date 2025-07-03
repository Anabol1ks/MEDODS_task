package response

type TokensResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

type UserResponse struct {
	UserID string `json:"user_id,omitempty"`
}

type LogoutResponse struct {
	Message string `json:"message,omitempty"`
}
