package middleware

import (
	"net/http"
	"strings"
	"time"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gin-gonic/gin"
)

var jwtSecretKey = []byte("daddi")

// CreateToken generates a JWT for a user with a given ID and role.
func CreateToken(userID uint, role string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "userID": userID,
        "role":   role,
        "exp":    time.Now().Add(time.Hour * 24).Unix(), // Token expiration: 24 hours
    })

    // Sign and get the complete encoded token as a string
    return token.SignedString(jwtSecretKey)
}

func AuthMiddleware(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        // Check if token is provided
        if tokenStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
            c.Abort()
            return
        }

        // Parse and validate the token
        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            // Validate the token signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return jwtSecretKey, nil
        })

        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
            c.Abort()
            return
        }

        // Check token claims
        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            // Extract role and userID from token claims
            role, roleOk := claims["role"].(string)

            // Check if the role is present and matches the required role
            if !roleOk || role != requiredRole {
                c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
                c.Abort()
                return
            }

            // If valid, set user ID for access in handlers
            c.Set("userID", claims["userID"])
            c.Set("role", role)
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Next()
    }
}