package services

import (
	"context"
	"errors"
	"user-api/models"
	"user-api/repository"
	"user-api/tracing"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// UserService handles business logic for user operations
type UserService struct {
	repo      repository.UserRepository
	validator *validator.Validate
	tracer    trace.Tracer
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo:      repo,
		validator: validator.New(),
		tracer:    tracing.GetTracer("user-api/services"),
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, s.tracer, "UserService.CreateUser")
	defer span.End()

	// Add request attributes
	tracing.AddSpanAttributes(span,
		tracing.AttrUserEmail.String(req.Email),
		attribute.String("user.first_name", req.FirstName),
		attribute.String("user.last_name", req.LastName),
	)

	// Validate the request
	tracing.AddSpanEvent(span, "validation.start")
	if err := s.validator.Struct(req); err != nil {
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("validation_error"))
		return nil, s.formatValidationError(err)
	}
	tracing.AddSpanEvent(span, "validation.success")

	// Check if user with email already exists
	tracing.AddSpanEvent(span, "email_check.start")
	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		err := errors.New("user with this email already exists")
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("duplicate_email"))
		return nil, err
	}
	tracing.AddSpanEvent(span, "email_check.success")

	// Create new user
	user := models.NewUser(req)
	tracing.AddSpanAttributes(span, tracing.AttrUserID.String(user.ID))

	// Save to repository
	tracing.AddSpanEvent(span, "repository.create.start")
	if err := s.repo.Create(ctx, user); err != nil {
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("repository_error"))
		return nil, err
	}
	tracing.AddSpanEvent(span, "repository.create.success")

	tracing.AddSpanAttributes(span, attribute.String("operation.result", "success"))
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, s.tracer, "UserService.GetUserByID")
	defer span.End()

	tracing.AddSpanAttributes(span, tracing.AttrUserID.String(id))

	if id == "" {
		err := errors.New("user ID is required")
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("validation_error"))
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("repository_error"))
		return nil, err
	}

	tracing.AddSpanAttributes(span,
		tracing.AttrUserEmail.String(user.Email),
		attribute.String("operation.result", "success"),
	)

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, s.tracer, "UserService.GetUserByEmail")
	defer span.End()

	tracing.AddSpanAttributes(span, tracing.AttrUserEmail.String(email))

	if email == "" {
		err := errors.New("email is required")
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("validation_error"))
		return nil, err
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("repository_error"))
		return nil, err
	}

	tracing.AddSpanAttributes(span,
		tracing.AttrUserID.String(user.ID),
		attribute.String("operation.result", "success"),
	)

	return user, nil
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, s.tracer, "UserService.GetAllUsers")
	defer span.End()

	users, err := s.repo.GetAll(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("repository_error"))
		return nil, err
	}

	tracing.AddSpanAttributes(span,
		attribute.Int("users.count", len(users)),
		attribute.String("operation.result", "success"),
	)

	return users, nil
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
