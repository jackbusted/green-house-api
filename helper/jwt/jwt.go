package jwt

import (
	"encoding/json"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

type JwtHelper struct {
	jwt.StandardClaims
}

func (self *JwtHelper) CreateJwtToken(secret string, ID string) (string, error) {
	claims := JwtHelper{}
	claims.Id = ID
	claims.ExpiresAt = time.Now().Add(300 * time.Minute).Unix()
	claims.ExpiresAt = time.Now().Add(30 * time.Hour).Unix()
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err := rawToken.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return token, err
}

func (self *JwtHelper) CreateJwtFreeToken(secret string, ID string) (string, error) {
	claims := JwtHelper{}
	jti, _ := json.Marshal(map[string]interface{}{
		"company_id":  0,
		"personal_id": 0,
		"user_id":     0,
	})
	ID = string(jti)
	claims.Id = ID
	claims.ExpiresAt = time.Now().Add(3 * time.Minute).Unix()
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err := rawToken.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return token, err
}

func (u *JwtHelper) GetJwtClaims(c echo.Context) jwt.MapClaims {
	user := c.Get("ID")
	if user == nil {
		return nil
	}

	token := user.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	return claims
}

func (u *JwtHelper) GetJwtClaim(c echo.Context, key string) interface{} {
	claims := u.GetJwtClaims(c)

	return claims[key]
}
