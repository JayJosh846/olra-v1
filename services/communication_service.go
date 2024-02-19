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
	"olra-v1/internal/structs"

	helpers "olra-v1/utils"

	"gopkg.in/gomail.v2"
)

var (
	termiiBaseURL = os.Getenv("TERMII_BASE_URL")
	termiiApiKey  = os.Getenv("TERMII_API_KEY")
)

func SendPhoneWelcomeOTP(mobile string) (string, error) {
	data := structs.PhoneOTPRequest{
		APIKey:         termiiApiKey,
		MessageType:    "NUMERIC",
		To:             mobile,
		From:           "N-Alert",
		Channel:        "dnd",
		PINAttempts:    3,
		PINTimeToLive:  10,
		PINLength:      6,
		PINPlaceholder: "< 1234 >",
		MessageText:    "Welcome to Olra! We are extremely glad to have you on our platform. Kindly use the confirmation code < 1234 > to verify your phone number. Valid for 10 minutes, one-time use only",
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

func SendPhoneOTP(mobile string) (string, error) {
	data := structs.PhoneOTPRequest{
		APIKey:         termiiApiKey,
		MessageType:    "NUMERIC",
		To:             mobile,
		From:           "N-Alert",
		Channel:        "dnd",
		PINAttempts:    3,
		PINTimeToLive:  10,
		PINLength:      6,
		PINPlaceholder: "< 1234 >",
		MessageText:    "Your Olra confirmation code is < 1234 >. Do not disclose this code with anyone. Valid for 10 minutes, one-time use only",
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

func SendRequestFundsSMS(mobile, firstName, lastName, tag string, amount float64) (string, error) {
	message := fmt.Sprintf(
		"Hello %s %s, you have received a request from %s to send the amount of %.2f to them. Kindly log into your Olra account to do so.",
		firstName, lastName, tag, amount,
	)

	data := structs.FundsRequest{
		APIKey:  termiiApiKey,
		To:      mobile,
		From:    "N-Alert",
		SMS:     message,
		Type:    "plain",
		Channel: "dnd",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", termiiBaseURL+"/sms/send", bytes.NewBuffer(jsonData))
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

func CreditFundsSMS(mobile, tag string, amount, newBalance float64) (string, error) {
	message := fmt.Sprintf(
		"Your Olra account has been credited with %.2f by %s. Your new account balance is %.2f",
		amount, tag, newBalance,
	)
	data := structs.FundsRequest{
		APIKey:  termiiApiKey,
		To:      mobile,
		From:    "N-Alert",
		SMS:     message,
		Type:    "plain",
		Channel: "dnd",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", termiiBaseURL+"/sms/send", bytes.NewBuffer(jsonData))
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

func DebitFundsSMS(mobile string, amount, newBalance float64) (string, error) {
	message := fmt.Sprintf(
		"Your Olra account has been debited with %.2f. Your new account balance is %.2f",
		amount, newBalance,
	)
	data := structs.FundsRequest{
		APIKey:  termiiApiKey,
		To:      mobile,
		From:    "N-Alert",
		SMS:     message,
		Type:    "plain",
		Channel: "dnd",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", termiiBaseURL+"/sms/send", bytes.NewBuffer(jsonData))
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
