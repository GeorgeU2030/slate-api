package dto

type UserRegisterRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	FullName *string `json:"fullName"`
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPairResponse struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken *string `json:"refreshToken,omitempty"`
}

type RefreshRequest struct {
	RefreshToken *string `json:"refreshToken"`
}

type GoogleLoginRequest struct {
	Code string `json:"code"`
}