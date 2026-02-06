package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"

	//"github.com/neo4j/neo4j-go-driver/neo4j/utils"
	"golang.org/x/crypto/bcrypt"
)

var Validate = validator.New()

type Error struct {
	StatusCode    int    `json:"statusCode"`
	Error         string `json:"error"`
	MessageToUser string `json:"messageToUser"`
}

func ParseBody(body io.Reader, out interface{}) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return err
	}

	return nil
}
func EncodeJSONBody(resp http.ResponseWriter, data interface{}) error {

	return json.NewEncoder(resp).Encode(data)
}

func RespondJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if body != nil {
		if err := EncodeJSONBody(w, body); err != nil {
			fmt.Println("Failed to respond JSON with error: %v", err)
		}
	}
}
func RespondError(w http.ResponseWriter, statusCode int, err error, messageToUser string) {
	w.WriteHeader(statusCode)
	var errString string
	if err != nil {
		errString = err.Error()
	}
	newError := Error{
		StatusCode:    statusCode,
		Error:         errString,
		MessageToUser: messageToUser,
	}
	if err := json.NewEncoder(w).Encode(newError); err != nil {
		fmt.Printf("failed to send error %v", err)
	}
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
func CheckPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(plainPassword),
	)
}
func ParseBool(str string) bool {
	if str == "" || str == "false" {
		return false
	}
	return true
}
func ParseExpiringAt(str string) (*time.Time, error) {
	var date time.Time
	if str != "" {
		d, err := time.Parse("2006-01-02", str)
		if err != nil {
			return nil, err
		}
		if d.Before(time.Now()) {
			return nil, errors.New("invalid time")
		}
		date = d
	}
	return &date, nil
}
func GenerateJWT(usrID, sessionID string) (string, error) {
	claims := jwt.MapClaims{
		"userId":    usrID,
		"sessionId": sessionID,
		"exp":       time.Now().Add(time.Minute * 100).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}
func GoDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv(key)
}
