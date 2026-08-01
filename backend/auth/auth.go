package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key, read from the JWT_SECRET environment variable.
// Falls back to a well-known dev-only value with a loud warning so it
// is never silently used in a real deployment.
var JwtKey = loadJwtKey()

func loadJwtKey() []byte {
	secret := db.GetEnv("JWT_SECRET", "")
	if secret == "" {
		log.Println("⚠️  WARNING: JWT_SECRET is not set. Using an insecure default key. " +
			"Set JWT_SECRET in the environment before deploying to production.")
		secret = "dev_only_insecure_default_key"
	}
	return []byte(secret)
}

// Claims for JWT
type Claims struct {
	UserID      int    `json:"userId"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	CompanyName string `json:"companyName"`
	jwt.RegisteredClaims
}

// AuthMiddleware authenticates JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":false,"message":"Authorization header is required. Please include a valid Bearer token in your request headers."}`))
			return
		}

		// Check if the header has the format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header format must be 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Store claims in context for further use
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "companyName", claims.CompanyName)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware restricts access based on user role
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (set by authMiddleware)
			role := r.Context().Value("role").(string)

			// Check if the role is allowed
			allowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GenerateToken creates a JWT token for a user
func GenerateToken(user models.User, roleName string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		CompanyName: user.CompanyName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}