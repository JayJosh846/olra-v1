package utils

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	mathrand "math/rand"
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
		return "", fmt.Errorf("invalid email format")
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

func GenerateRandomAccountNumber() string {
	// Generate the first two digits starting with "00"
	firstTwoDigits := "00"

	// Generate the remaining eight digits randomly
	var remainingDigits string
	for i := 0; i < 8; i++ {
		remainingDigits += fmt.Sprintf("%d", mathrand.Intn(10))
	}
	return firstTwoDigits + remainingDigits
}

func ValidateTagRequest(tagRequest string) error {
	// Regular expression to match only alphabets and numbers
	regex := regexp.MustCompile("^[a-z]+[a-z0-9]*$")

	// Check if the tag contains only alphabets and numbers
	if !regex.MatchString(tagRequest) {
		return errors.New("tag can only contain lowercase alphabets and numbers, and must start with an alphabet")
	}

	return nil
}

func ValidateOnlyNumbers(input string) bool {
	// Regular expression to match exactly 6 digits
	regex := regexp.MustCompile(`^\d{6}$`)
	return regex.MatchString(input)
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

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := cryptorand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GenerateRandomString(s int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes, err := generateRandomBytes(s)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
