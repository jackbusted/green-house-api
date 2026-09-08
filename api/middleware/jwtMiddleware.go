package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"green-house-api/helper/cache"
	"green-house-api/helper/logger"
	"green-house-api/model"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

type (
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Signing key to validate token.
		// Required.
		SigningKey interface{}

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Optional. Default value jwt.MapClaims
		Claims jwt.Claims

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup string

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		keyFunc  jwt.Keyfunc
		DBMaster *gorm.DB
	}

	jwtExtractor func(echo.Context) (string, error)
)

const (
	AlgorithmHS256 = "HS256"
)

var (
	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		Skipper:       middleware.DefaultSkipper,
		SigningMethod: AlgorithmHS256,
		ContextKey:    "ID",
		TokenLookup:   "header:" + echo.HeaderAuthorization,
		AuthScheme:    "Bearer",
		Claims:        jwt.MapClaims{},
	}
)

func JWT(key interface{}) echo.MiddlewareFunc {
	c := DefaultJWTConfig
	c.SigningKey = key
	return JWTWithConfig(c)
}

func JWTWithConfig(config JWTConfig) echo.MiddlewareFunc {
	logger.Default().Println("start of JWTWithConfig")
	defer logger.Default().Println("end of JWTWithConfig")

	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}
	if config.SigningKey == nil {
		panic("echo: jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultJWTConfig.AuthScheme
	}
	config.keyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
		}
		return config.SigningKey, nil
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := jwtFromHeader(parts[1], config.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = jwtFromQuery(parts[1])
	case "cookie":
		extractor = jwtFromCookie(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger.Default().Println("start of JWTWithConfig 2")
			defer logger.Default().Println("end of JWTWithConfig 2")
			//skip endpoint
			if strings.Contains(c.Request().RequestURI, "login") ||
				strings.Contains(c.Request().RequestURI, "forgot-password") ||
				strings.Contains(c.Request().RequestURI, "authentication") ||
				strings.Contains(c.Request().RequestURI, "authorization") ||
				strings.Contains(c.Request().RequestURI, "settings") ||
				strings.Contains(c.Request().RequestURI, "/callback") {
				return next(c)
			}

			auth, err := extractor(c)

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    http.StatusUnauthorized,
					"status":  "error",
					"message": err.Error(),
				})
			}
			token := new(jwt.Token)

			// Issue #647, #656
			if _, ok := config.Claims.(jwt.MapClaims); ok {
				token, err = jwt.Parse(auth, config.keyFunc)
			} else {
				claims := reflect.ValueOf(config.Claims).Interface().(jwt.Claims)
				token, err = jwt.ParseWithClaims(auth, claims, config.keyFunc)
			}

			if err == nil && token.Valid {
				var jti map[string]interface{}
				mapClaims := token.Claims.(jwt.MapClaims)
				err := json.Unmarshal([]byte(mapClaims["jti"].(string)), &jti)

				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"code":    http.StatusInternalServerError,
						"status":  "error",
						"message": err.Error(),
					})
				} else {
					var personal model.PersonalModel
					var user model.UserModel
					if c.Request().Header.Get("USER-SOURCE") == "Guest" {
						log.Println("USER-SOURCE = Guest")
						c.Set("user", model.UserModel{})
						c.Set("personal_id", uint(0))
						log.Println("masuk sini 2")
						return next(c)
					}

					cacheKey := fmt.Sprintf("personal:id=%v", uint(jti["personal_id"].(float64)))
					err = cache.GetCache(cacheKey, &personal)
					if err != nil {
						err = fmt.Errorf("get cache %v : %v", cacheKey, err.Error())
						logger.Default().Println(err.Error())
						return c.JSON(http.StatusUnauthorized, map[string]interface{}{
							"code":    http.StatusUnauthorized,
							"status":  "error",
							"message": "get session cache not found",
						})
					}

					if personal.IsActive != "1" {
						return c.JSON(http.StatusUnauthorized, map[string]interface{}{
							"code":    http.StatusUnauthorized,
							"status":  "error",
							"message": "Your account is inactive, please contact your administrator",
						})
					}

					cacheKey = fmt.Sprintf("user:id=%v", personal.UserID)
					err = cache.GetCache(cacheKey, &user)
					if err != nil {
						err = fmt.Errorf("get cache %v : %v", cacheKey, err.Error())
						logger.Default().Println(err.Error())
						return c.JSON(http.StatusUnauthorized, map[string]interface{}{
							"code":    http.StatusUnauthorized,
							"status":  "error",
							"message": "get session cache not found",
						})
					}

					c.Set("user", user)
					c.Set("personal", personal)
					c.Set("personal_id", personal.ID)
					c.Set("user_id", personal.UserID)
					c.Set("is_customer", false)
					c.Set("is_sales", true)

					return next(c)
				}
			}

			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    http.StatusUnauthorized,
				"status":  "error",
				"message": err.Error(),
			})
		}
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", errors.New("Missing or invalid jwt in the request header")
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(param string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", errors.New("Missing jwt in the query string")
		}
		return token, nil
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the named cookie.
func jwtFromCookie(name string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			return "", errors.New("Missing jwt in the cookie")
		}
		return cookie.Value, nil
	}
}
