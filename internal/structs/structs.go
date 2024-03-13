package structs

type VerifyPhoneOTPRequestBody struct {
	PinId string `json:"pinId" validate:"required"`
	Pin   string `json:"pin" validate:"required"`
	Phone string `json:"phone" validate:"required"`
}

type VerifyPhoneOTPBody struct {
	PinId    string `json:"pinId"`
	Verified bool   `json:"verified"`
	// Verifiedd string `json:"verified"`
	Msisdn string `json:"msisdn"`
	Status int    `json:"status"`
}

type PhoneOTPRequestBody struct {
	Mobile string `json:"mobile" validate:"required"`
}

type PhoneOTPResponse struct {
	PinId     string `json:"pinId"`
	To        string `json:"to"`
	SmsStatus string `json:"smsStatus"`
	Status    int    `json:"status"`
}

type UserRequestBody struct {
	FirstName   string `json:"firstName" validate:"required"`
	LastName    string `json:"lastName" validate:"required"`
	Email       string `json:"email" validate:"required"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type EmailVerificationCodeRequest struct {
	Code  string `json:"code"`
	Email string `json:"email" validate:"required"`
}

type BVNRequest struct {
	Bvn string `json:"bvn"`
}

type VerifyBVNResponse struct {
	Status  bool                    `json:"status"`
	Message string                  `json:"message"`
	Data    VerifyBVNResponseEntity `json:"data"`
}

type VerifyBVNResponseEntity struct {
	Entity VerifyBVNResponseBvn `json:"entity"`
}

type VerifyBVNResponseBvn struct {
	Bvn VerifyBVNResponseData `json:"bvn"`
}

type VerifyBVNResponseData struct {
	Status bool `json:"status"`
}

type TagRequest struct {
	Tag string `json:"tag"`
}

type PasscodeRequest struct {
	Passcode        string `json:"passcode"`
	ConfirmPasscode string `json:"confirmPasscode"`
}

type CallbackData struct {
	Title   string `json:"Title"`
	Message string `json:"Message"`
	Data    Data   `json:"Data"`
}

type UsersTags struct {
	User_ID uint   `json:"user_id"`
	Tag     string `json:"tag"`
}

// Data represents the dynamic data structure within the callback data
type Data struct {
	NUBANName   string `json:"NUBANName"`
	NUBAN       string `json:"NUBAN"`
	NUBANStatus string `json:"NUBANStatus"`
	NUBANType   int    `json:"NUBANType"`
	Request     int    `json:"Request"`
}

type WalletRequest struct {
	Gender      string `json:"gender"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Dob         string `json:"dob"`
	PhoneNumber string `json:"phoneNumber"`
}

type GenerateWalletResponse struct {
	Successful bool   `json:"Successful"`
	Message    string `json:"Message"`
}

type LoginRequestBody struct {
	PhoneNumber string `json:"phoneNumber"`
	Passcode    string `json:"passcode"`
	DeviceID    string `json:"deviceId"`
}

type RequestFundsRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Requestee   string  `json:"requestee"`
}

type SendFundsRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Receiver    string  `json:"receiver"`
}
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

type FundsRequest struct {
	APIKey  string `json:"api_key"`
	To      string `json:"to"`
	From    string `json:"from"`
	SMS     string `json:"sms"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

type FundsResponse struct {
	Message_ID string  `json:"message_id"`
	Message    string  `json:"message"`
	Balance    float64 `json:"balance"`
	User       string  `json:"user"`
}

type GroupRequest struct {
	GroupName    string                `json:"groupName"`
	Tag          string                `json:"tag"`
	Amount       int                   `json:"amount"`
	GroupMembers []GroupFriendsRequest `json:"groupMembers"`
}

type GroupFriendsRequest struct {
	Friend string `json:"friend"`
}

type SendGroupFundsRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Group       string  `json:"group"`
}

type WaitlistRequest struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type ContactRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Message   string `json:"message"`
}
