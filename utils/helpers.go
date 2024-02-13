package utils

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	mathrand "math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Usererror struct {
	Error        bool   `json:"error"`
	ResponseCode int    `json:"response code"`
	Message      string `json:"message"`
	Data         string `json:"data"`
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userpassword string, givenpassword string) (bool, Usererror) {
	err := bcrypt.CompareHashAndPassword([]byte(givenpassword), []byte(userpassword))
	valid := true
	if err != nil {
		valid = false
		return valid,
			Usererror{
				Error:        true,
				ResponseCode: 400,
				Message:      "Invalid Password",
				Data:         "",
			}
	}
	return valid, Usererror{}
}

func GenerateTransactionReference() string {
	// Generate a random identifier.
	identifier := mathrand.Intn(1000000) // Change the range as needed.
	// Get the current timestamp.
	currentTime := time.Now()
	// Format the timestamp and combine it with the identifier.
	transactionReference := currentTime.Format("20060102150405") + fmt.Sprintf("%06d", identifier)
	return transactionReference
}

func ExtractUsernameFromEmail(email string) (string, error) {
	// Remove special characters (except '@' and '.') from the email
	re := regexp.MustCompile("[^a-zA-Z0-9@.-]+")
	cleanedEmail := re.ReplaceAllString(email, "")

	// Remove dot (.) if it exists before '@'
	parts := strings.Split(cleanedEmail, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid email format")
	}
	username := strings.ReplaceAll(parts[0], ".", "")

	return username, nil
}

func GenerateRandomPassword(length int) (string, error) {
	numBytes := (length * 3) / 4

	// Generate random bytes
	randomBytes := make([]byte, numBytes)
	_, err := cryptorand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	password := base64.URLEncoding.EncodeToString(randomBytes)
	password = password[:length]

	return password, nil
}

func GenerateVerificationCode() string {
	code := mathrand.Intn(900000) + 100000
	return strconv.Itoa(code)
}

func GetGoogleUserInfo(accessToken string) (map[string]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func GetFacebookUserInfo(accessToken string) (map[string]interface{}, error) {
	facebookRequestURL := fmt.Sprintf("https://graph.facebook.com/v19.0/me?fields=id%2Cname&access_token=%s", accessToken)
	resp, err := http.Get(facebookRequestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return userInfo, nil
}

func MapToStruct(inputMap map[string]interface{}, resultStruct interface{}) (interface{}, error) {
	resultValue := reflect.ValueOf(resultStruct).Elem()
	for key, value := range inputMap {
		field := resultValue.FieldByName(key)
		if field.IsValid() {
			if field.CanSet() {
				fieldType := field.Type()
				if reflect.TypeOf(value).ConvertibleTo(fieldType) {
					field.Set(reflect.ValueOf(value).Convert(fieldType))
				}
			}
		}
	}
	return resultStruct, nil
}
