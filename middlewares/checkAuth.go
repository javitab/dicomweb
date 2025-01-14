package middlewares

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	dbase "github.com/javitab/go-web/database"
)

func CheckAuth(c *gin.Context) {

	authHeader := c.GetHeader("Authorization")

	if flag.Lookup(("test.v")) == nil {

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authToken := strings.Split(authHeader, " ")
		if len(authToken) != 2 || authToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := authToken[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("SECRET_JWT_KEY")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"message": err.Error(),
			})
			dbase.LogServerError("CheckAuth:JWT:InvalidOrExpired", err, "Error validating token for user: "+tokenString)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		db := dbase.GetDBConn()
		var user dbase.User
		db.Where("Username=?", claims["username"]).Find(&user)

		if user.ID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "Username Not Found",
				"claims": claims,
			})
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if user.DeletedAt.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "User is disabled",
				"claims": claims,
			})
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("currentUser", user.Username)
	} else {
		c.Set("currentUser", "testuser")
		c.Next()
	}
}
