package models

// AuthRegisterRequest digunakan untuk request register user baru
type AuthRegisterRequest struct {
	Username string `json:"username" example:"example"`
	Email    string `json:"email" example:"example@mail.com"`
	Password string `json:"password" example:"123456"`
}

// AuthLoginRequest digunakan untuk request login
type AuthLoginRequest struct {
	Email    string `json:"email" example:"example@mail.com"`
	Password string `json:"password" example:"123456"`
}

// AuthResponse digunakan untuk response login/register
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse digunakan untuk response user
type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UserUpdateRequest digunakan untuk update data user
type UserUpdateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DeleteResponse digunakan untuk response delete endpoint
type DeleteResponse struct {
	Message string `json:"message"`
}
