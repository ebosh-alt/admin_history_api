package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"admin_history/internal/authctx"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JWTAuthMiddleware возвращает gin.HandlerFunc, который проверяет
// JWT access-токен в заголовке Authorization и кладёт userID в контекст.
// При ошибке возвращает 401 Unauthorized.
// jwtMgr — ваш auth.JWTManager,
// logger — zap.Logger для логов.
func (m *Middleware) JWTAuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		m.log.Warn("JWT auth failed: missing or malformed Authorization header",
			zap.String("header", authHeader),
		)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "missing or malformed Authorization header",
		})
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := m.VerifyAccessToken(tokenStr)
	if err != nil {
		m.log.Warn("JWT auth failed: invalid or expired token", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid or expired token",
		})
		return
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		m.log.Warn("JWT verify: bad user id in claims", zap.String("uid", claims.UserID))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
		return
	}

	// 1) как раньше: можно оставить для дебага/удобства в gin
	c.Set("userID", claims.UserID)

	// 2) ГЛАВНОЕ: положить в request.Context()
	ctx := authctx.WithUserID(c.Request.Context(), uid)
	// (опционально) если держишь роль в клеймах:
	isAdmin := claims.Role == "admin" || claims.Role == "super_admin"
	ctx = authctx.WithIsAdmin(ctx, isAdmin)

	c.Request = c.Request.WithContext(ctx)

	m.log.Debug("JWT auth success", zap.String("userID", claims.UserID), zap.Bool("isAdmin", isAdmin))
	c.Next()
}

type AccessClaims struct {
	jwt.RegisteredClaims        // sub, exp, iat, nbf, iss, aud, jti
	UserID               string `json:"uid"` // наш кастомный ID
	Role                 string `json:"role,omitempty"`
}

func GenerateTokens(userID uuid.UUID, role string, AccessTTL time.Duration, RefreshTTL time.Duration, Secret string) (string, string, time.Time, time.Time, error) {
	now := time.Now()

	// 1) Access-JWT
	accessClaims := AccessClaims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTTL)),
		},
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := at.SignedString([]byte(Secret))
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	// 2) Refresh (random string)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(b)
	refreshExp := now.Add(RefreshTTL)

	return accessToken, refreshToken, now.Add(AccessTTL), refreshExp, nil
}

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidTokenClaims = errors.New("invalid token claims")
)

func (m *Middleware) VerifyAccessToken(tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}

	keyFunc := func(t *jwt.Token) (any, error) {
		// Жёстко проверяем метод подписи
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(m.cfg.JWT.Secret), nil
	}

	// Параноидальные опции валидации
	tkn, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		keyFunc,
		jwt.WithLeeway(30*time.Second),          // небольшой лейвей на часы
		jwt.WithIssuedAt(),                      // проверять iat (если есть)
		jwt.WithValidMethods([]string{"HS256"}), // валидные методы
		// при необходимости можно включить:
		// jwt.WithAudience(m.cfg.JWT.Aud),
		// jwt.WithIssuer(m.cfg.JWT.Iss),
	)
	if err != nil {
		m.log.Warn("JWT verify: parse failed", zap.Error(err))
		return nil, ErrInvalidToken
	}
	if !tkn.Valid {
		m.log.Warn("JWT verify: token invalid")
		return nil, ErrInvalidToken
	}

	if claims.UserID == "" && claims.Subject != "" {
		claims.UserID = claims.Subject
	}

	// (Необязательно) последний шанс: посмотреть в map claims, если в проде уже гуляют токены с другим именем поля, напр. "user_id"
	if claims.UserID == "" {
		if mc, ok := tkn.Claims.(jwt.MapClaims); ok {
			if v, ok := mc["user_id"].(string); ok && v != "" {
				claims.UserID = v
			}
		}
	}

	if claims.UserID == "" {
		m.log.Warn("JWT verify: empty user id in claims", zap.Any("claims", claims))
		return nil, ErrInvalidTokenClaims
	}

	return claims, nil
}
