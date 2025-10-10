package models

type CheckPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}
