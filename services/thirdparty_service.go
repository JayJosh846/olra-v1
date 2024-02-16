package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type BvnRequest struct {
	Number string `json:"number"`
	// Type string `json:"type"`

}

type WalletRequest struct {
	Gender      string `json:"gender"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Dob         string `json:"dob"`
	PhoneNumber string `json:"phoneNumber"`
}

func VerifyBVN(number string) (string, error) {
	apiKey := os.Getenv("IDENTITY_API_KEY")
	apiID := os.Getenv("IDENTITY_APP_ID")
	//	secKey := os.Getenv("CHECKID_SEC_KEY")
	// token := secKey
	// typeBody := "validate"
	// 	url := "https://sandbox.checkid.ng/api/v1/identity/bvn"
	url := "https://api.myidentitypay.com/api/v2/biometrics/merchant/data/verification/bvn"
	method := "POST"
	bvnRequest := BvnRequest{
		Number: number,
		// Type: typeBody,
	}
	requestBodyJSON, err := json.Marshal(bvnRequest)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return "", err
	}
	bodyReader := bytes.NewReader([]byte(requestBodyJSON))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	//	req.Header.Add("Authorization", "Bearer "+token)

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("app-id", apiID)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("body: ", string(body))
	return string(body), nil
}

func GenerateWallet(gender, email, firstName, lastName, dob, phoneNumber string) (string, error) {
	url := "https://wema-alatdev-apimgt.azure-api.net/onboarding-wallets/api/CustomerAccount/GenerateWalletV2"
	method := "POST"
	bvnRequest := WalletRequest{
		Gender:      gender,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Dob:         dob,
		PhoneNumber: phoneNumber,
	}
	requestBodyJSON, err := json.Marshal(bvnRequest)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return "", err
	}
	bodyReader := bytes.NewReader([]byte(requestBodyJSON))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	//	req.Header.Add("Authorization", "Bearer "+token)

	req.Header.Set("Ocp-Apim-Subscription-Key", "8ae5210581d04e589db7c1cded1e80fb")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("body: ", string(body))
	return string(body), nil
}

func Callback() (string, error) {
	// secKey := os.Getenv("PAYSTACK_SEC_KEY")
	// token := secKey
	url := "https://wema-alatdev-apimgt.azure-api.net/callback-url/callbackURL"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", "8ae5210581d04e589db7c1cded1e80fb")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()
	// Read the response body
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}
	return string(responseBody), nil

}
