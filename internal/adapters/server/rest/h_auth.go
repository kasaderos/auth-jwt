package rest

import (
	"birthday-bot/internal/domain/entities"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var tokenTTL = 30 * time.Minute

// Create the JWT key used to create the signature
var jwtKey = []byte("auth-jwt-1234")

// @Router   /register [post]
// @Tags     register
// @Param    body  body  entities.UserSt false  "body"
// @Success  200  {object}
// @Failure  400  {object}  dopTypes.ErrRep
func (o *St) hRegister(c *gin.Context) {
	var user entities.UserCUSt
	if !BindJSON(c, user) {
		return
	}

	// TODO validate

	result, err := o.ucs.UserCreate(o.getRequestContext(c), &user)
	if Error(c, err) {
		return
	}

	// create activation jwt token
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &entities.Claims{
		Login: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// send url
	// TODO BaseURL
	message := fmt.Sprintf("%s/activate?token=%s", "http://localhost:8080", token)

	err = o.ucs.SendMessage(user.Email, message)
	if err != nil {
		// TODO resend or put logic into mail.Client
		return
	}
}

// @Router   /auth [post]
// @Tags     auth
// @Param    body  body  entities.Credentials false  "body"
// @Success  200  {object}
// @Failure  400  {object}  dopTypes.ErrRep
func (o *St) hAuth(c *gin.Context) {
	var creds entities.Credentials
	if !BindJSON(c, creds) {
		return
	}

	// TODO validate

	user, err := o.ucs.GetUserByEmail(o.getRequestContext(c), creds.Login)
	if Error(c, err) || user.Password != creds.Password {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Declare the expiration time of the token
	// here, we have kept it as 30 minutes
	expirationTime := time.Now().Add(tokenTTL)
	// Create the JWT claims, which includes the username and expiry time
	claims := &entities.Claims{
		Login: creds.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func (o *St) hRefresh(c *gin.Context) {
	c, err := c.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	tknStr := c.Value
	claims := &entities.Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	if time.Until(claims.ExpiresAt.Time) > 30*time.Second {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `token` cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func (o *St) hProfile(c *gin.Context) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := c.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			c.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &entities.Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := o.ucs.GetUserByEmail(o.getRequestContext(c), creds.Login)
	if Error(c, err) {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (o *St) hLogout(c *gin.Context) {
	// immediately clear the token cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
	})
}
