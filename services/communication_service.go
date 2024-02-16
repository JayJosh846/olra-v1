package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"olra-v1/internal/database"
	helpers "olra-v1/utils"

	"gopkg.in/gomail.v2"
)

var (
	termiiBaseURL = os.Getenv("TERMII_BASE_URL")
	termiiApiKey  = os.Getenv("TERMII_API_KEY")
)

type PhoneOTPRequest struct {
	APIKey         string `json:"api_key"`
	MessageType    string `json:"message_type"`
	To             string `json:"to"`
	From           string `json:"from"`
	Channel        string `json:"channel"`
	PINAttempts    int    `json:"pin_attempts"`
	PINTimeToLive  int    `json:"pin_time_to_live"`
	PINLength      int    `json:"pin_length"`
	PINPlaceholder string `json:"pin_placeholder"`
	MessageText    string `json:"message_text"`
	PINType        string `json:"pin_type"`
}

func SendPhoneOTP(mobile string) (string, error) {
	data := PhoneOTPRequest{
		APIKey:         termiiApiKey,
		MessageType:    "NUMERIC",
		To:             mobile,
		From:           "N-Alert",
		Channel:        "dnd",
		PINAttempts:    3,
		PINTimeToLive:  10,
		PINLength:      6,
		PINPlaceholder: "< 1234 >",
		MessageText:    "Hello. Your Olra confirmation code is < 1234 >. Valid for 10 minutes, one-time use only",
		PINType:        "NUMERIC",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", termiiBaseURL+"/sms/otp/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("cache-control", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func VerifyOTP(pinID, pin string) (string, error) {
	data := map[string]string{
		"api_key": termiiApiKey,
		"pin_id":  pinID,
		"pin":     pin,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", termiiBaseURL+"/sms/otp/verify", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("cache-control", "no-cache")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func CreateEmailOtp(user *database.User, email string) (*database.Otp, error) {

	// Generate verification code
	verificationCode := helpers.GenerateVerificationCode()
	// Set expiration time
	expirationTime := time.Now().Add(30 * time.Minute)

	// Create new OTP record
	otp := &database.Otp{
		UserID:    user.UserID,
		Token:     verificationCode,
		ExpiresAt: expirationTime,
	}
	if err := database.DB.Create(otp).Error; err != nil {
		return nil, err
	}

	// Send verification email
	//  email := user.Email // Assuming email is stored in the User model
	if err := sendEmailOTP(user.FirstName, email, verificationCode); err != nil {
		return nil, err
	}
	return otp, nil
}

func sendEmailOTP(firstName, email, code string) error {

	m := gomail.NewMessage()
	m.SetHeader("From", "jesudara@withpepp.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "OTP Code")
	m.SetBody("text/plain", fmt.Sprintf("Hello %s,\n\nYour otp code is: %s, valid for 30mins.", firstName, code))

	d := gomail.NewDialer("smtp.zoho.com", 465, "jesudara@withpepp.com", "143Asdf846@")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Println("Email not sent.")
		return err
	} else {
		log.Println("Email sent successfully.")
	}

	return nil
}
