package tests

// import (
// 	"bytes"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"

// 	"github.com/gin-gonic/gin"
// )

// func TestHelloWorldHandler(t *testing.T) {
// 	s := &server.Server{}
// 	r := gin.New()
// 	r.GET("/", s.HelloWorldHandler)
// 	// Create a test HTTP request
// 	req, err := http.NewRequest("GET", "/", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// Create a ResponseRecorder to record the response
// 	rr := httptest.NewRecorder()
// 	// Serve the HTTP request
// 	r.ServeHTTP(rr, req)
// 	// Check the status code
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 	}
// 	// Check the response body
// 	expected := "{\"message\":\"Hello World\"}"
// 	if rr.Body.String() != expected {
// 		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
// 	}
// }

// func TestRequestPhoneOTP(t *testing.T) {
// 	// Initialize your Gin router
// 	router := gin.New()

// 	// Define a JSON payload for the request
// 	jsonPayload := `{"mobile": "+2347040247157"}`

// 	// Create a request with the JSON payload
// 	req, err := http.NewRequest("POST", "http://localhost:9090/api/v1/user/send-phone-otp", bytes.NewBuffer([]byte(jsonPayload)))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	// Create a ResponseRecorder to record the response
// 	rr := httptest.NewRecorder()

// 	// Serve the HTTP request to the endpoint
// 	router.ServeHTTP(rr, req)

// 	// Check the status code
// 	assert.Equal(t, http.StatusOK, rr.Code)

// 	// Further assertions can be made on the response body, headers, etc.
// 	assert.Contains(t, rr.Body.String(), "OTP sent successfully")
// }
