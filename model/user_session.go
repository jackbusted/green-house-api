package model

type UserSessionModel struct {
	DefaultAttribute
	UserID     uint
	ClientID   uint
	Token      string
	ExpiryTime int64
}

func (UserSessionModel) TableName() string {
	return "user_sessions"
}
