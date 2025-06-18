package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name" validate:"required,min=2,max=50"`
	LastName    string    `json:"last_name" validate:"required,min=2,max=50"`
	Email       string    `json:"email" validate:"required,email"`
	Phone       string    `json:"phone,omitempty" validate:"omitempty,min=10,max=15"`
	DateOfBirth string    `json:"date_of_birth,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Address     *Address  `json:"address,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Address represents a user's address
type Address struct {
	Street     string `json:"street,omitempty" validate:"omitempty,max=100"`
	City       string `json:"city,omitempty" validate:"omitempty,max=50"`
	State      string `json:"state,omitempty" validate:"omitempty,max=50"`
	PostalCode string `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	Country    string `json:"country,omitempty" validate:"omitempty,max=50"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	FirstName   string   `json:"first_name" validate:"required,min=2,max=50"`
	LastName    string   `json:"last_name" validate:"required,min=2,max=50"`
	Email       string   `json:"email" validate:"required,email"`
	Phone       string   `json:"phone,omitempty" validate:"omitempty,min=10,max=15"`
	DateOfBirth string   `json:"date_of_birth,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Address     *Address `json:"address,omitempty"`
}

// NewUser creates a new user from a create request
func NewUser(req CreateUserRequest) *User {
	now := time.Now()
	return &User{
		ID:          uuid.New().String(),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		DateOfBirth: req.DateOfBirth,
		Address:     req.Address,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// UserResponse represents the response format for user data
type UserResponse struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone,omitempty"`
	DateOfBirth string    `json:"date_of_birth,omitempty"`
	Address     *Address  `json:"address,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		FullName:    u.GetFullName(),
		Email:       u.Email,
		Phone:       u.Phone,
		DateOfBirth: u.DateOfBirth,
		Address:     u.Address,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
