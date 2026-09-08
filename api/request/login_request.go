package request

type Credentials struct {
	Email              string `json:"email" validate:"required"`
	Password           string `json:"password" validate:"required"`
	AuthenticationID   string `json:"authentication_id"`
	AuthenticationType string `json:"authentication_type"`
	Code               string `json:"code"`
	SourceAppID        string `json:"source_app_id"`
}
