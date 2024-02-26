package controllers

import (
	"auth/initializers"
	"auth/models"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	//Get the email/pass from req body
	var body struct {
		Email    string
		Password string
	}

	//Bind the body
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})

		return
	}

	// Validate password
    if len(body.Password) < 8 || len(body.Password) > 12 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8-12 characters long"})
        return
    }

    // Check if password contains at least one letter and one number
    matched, _ := regexp.MatchString(`(?i)[a-z]+`, body.Password)
    if !matched {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Password must contain at least one letter"})
        return
    }
    matched, _ = regexp.MatchString(`\d+`, body.Password)
    if !matched {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Password must contain at least one number"})
        return
    }

	//Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
	}

	//Create the user
	user := models.User{Email: body.Email, Password: string(hash)}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user"})
		return
	}

	//Respond
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})

}

func Login(c *gin.Context) {
	//Get the email/pass from req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})

		return
	}

	//Look up requested user
	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}

	//Compare sent in pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}
	//Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	//Sign and get the complete encoded token as a string using the secret key
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to generate token"})
		return
	}

	//send it back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600 * 24 * 30, "","", false, true)
	c.JSON(http.StatusOK, gin.H{})

}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})
}
