package services

import (
	"errors"
	"user-api/models"
	"user-api/repository"

	"github.com/go-playground/validator/v10"
)

// UserService handles business logic for user operations
type UserService struct {
	repo      repository.UserRepository
	validator *validator.Validate
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo:      repo,
		validator: validator.New(),
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req models.CreateUserRequest) (*models.User, error) {
	// Validate the request
	if err := s.validator.Struct(req); err != nil {
		return nil, s.formatValidationError(err)
	}

	// Check if user with email already exists
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create new user
	user := models.NewUser(req)

	// Save to repository
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}

	return s.repo.GetByID(id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	return s.repo.GetByEmail(email)
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.repo.GetAll()
}

// formatValidationError formats validation errors into a readable message
func (s *UserService) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorMessages []string
		for _, fieldError := range validationErrors {
			switch fieldError.Tag() {
			case "required":
				errorMessages = append(errorMessages, fieldError.Field()+" is required")
			case "email":
				errorMessages = append(errorMessages, fieldError.Field()+" must be a valid email address")
			case "min":
				errorMessages = append(errorMessages, fieldError.Field()+" must be at least "+fieldError.Param()+" characters long")
			case "max":
				errorMessages = append(errorMessages, fieldError.Field()+" must be at most "+fieldError.Param()+" characters long")
			case "datetime":
				errorMessages = append(errorMessages, fieldError.Field()+" must be in YYYY-MM-DD format")
			default:
				errorMessages = append(errorMessages, fieldError.Field()+" is invalid")
			}
		}

		// Join all error messages
		var combinedMessage string
		for i, msg := range errorMessages {
			if i > 0 {
				combinedMessage += "; "
			}
			combinedMessage += msg
		}
		return errors.New(combinedMessage)
	}

	return err
}
