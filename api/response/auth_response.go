package response

type DoAuthentication struct {
	Token      string `json:"token"`
	LogLoginID uint   `json:"log_login_id"`
}
