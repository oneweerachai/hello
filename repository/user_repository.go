package repository

import (
	"context"
	"errors"
	"sync"
	"user-api/models"
	"user-api/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
}

// InMemoryUserRepository implements UserRepository using in-memory storage
type InMemoryUserRepository struct {
	users  map[string]*models.User
	mutex  sync.RWMutex
	tracer trace.Tracer
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[string]*models.User),
		mutex:  sync.RWMutex{},
		tracer: tracing.GetTracer("user-api/repository"),
	}
}

// Create adds a new user to the repository
func (r *InMemoryUserRepository) Create(ctx context.Context, user *models.User) error {
	ctx, span := tracing.StartSpan(ctx, r.tracer, "InMemoryUserRepository.Create")
	defer span.End()

	tracing.AddSpanAttributes(span,
		tracing.AttrDBOperation.String("create"),
		tracing.AttrDBTable.String("users"),
		tracing.AttrUserID.String(user.ID),
		tracing.AttrUserEmail.String(user.Email),
	)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user with same email already exists
	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			err := errors.New("user with this email already exists")
			tracing.RecordError(span, err)
			tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("duplicate_email"))
			return err
		}
	}

	r.users[user.ID] = user
	tracing.AddSpanAttributes(span, attribute.String("operation.result", "success"))
	return nil
}

// GetByID retrieves a user by ID
func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, r.tracer, "InMemoryUserRepository.GetByID")
	defer span.End()

	tracing.AddSpanAttributes(span,
		tracing.AttrDBOperation.String("get_by_id"),
		tracing.AttrDBTable.String("users"),
		tracing.AttrUserID.String(id),
	)

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		err := errors.New("user not found")
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("not_found"))
		return nil, err
	}

	tracing.AddSpanAttributes(span,
		tracing.AttrUserEmail.String(user.Email),
		attribute.String("operation.result", "success"),
	)
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *InMemoryUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, r.tracer, "InMemoryUserRepository.GetByEmail")
	defer span.End()

	tracing.AddSpanAttributes(span,
		tracing.AttrDBOperation.String("get_by_email"),
		tracing.AttrDBTable.String("users"),
		tracing.AttrUserEmail.String(email),
	)

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			tracing.AddSpanAttributes(span,
				tracing.AttrUserID.String(user.ID),
				attribute.String("operation.result", "success"),
			)
			return user, nil
		}
	}

	err := errors.New("user not found")
	tracing.RecordError(span, err)
	tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("not_found"))
	return nil, err
}

// GetAll retrieves all users
func (r *InMemoryUserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, r.tracer, "InMemoryUserRepository.GetAll")
	defer span.End()

	tracing.AddSpanAttributes(span,
		tracing.AttrDBOperation.String("get_all"),
		tracing.AttrDBTable.String("users"),
	)

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	tracing.AddSpanAttributes(span,
		attribute.Int("users.count", len(users)),
		attribute.String("operation.result", "success"),
	)
	return users, nil
}

// Update updates an existing user
func (r *InMemoryUserRepository) Update(ctx context.Context, user *models.User) error {
	ctx, span := tracing.StartSpan(ctx, r.tracer, "InMemoryUserRepository.Update")
	defer span.End()

	tracing.AddSpanAttributes(span,
		tracing.AttrDBOperation.String("update"),
		tracing.AttrDBTable.String("users"),
		tracing.AttrUserID.String(user.ID),
		tracing.AttrUserEmail.String(user.Email),
	)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		err := errors.New("user not found")
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("not_found"))
		return err
	}

	r.users[user.ID] = user
	tracing.AddSpanAttributes(span, attribute.String("operation.result", "success"))
	return nil
}

// Delete removes a user from the repository
func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	ctx, span := tracing.StartSpan(ctx, r.tracer, "InMemoryUserRepository.Delete")
	defer span.End()

	tracing.AddSpanAttributes(span,
		tracing.AttrDBOperation.String("delete"),
		tracing.AttrDBTable.String("users"),
		tracing.AttrUserID.String(id),
	)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[id]; !exists {
		err := errors.New("user not found")
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("not_found"))
		return err
	}

	delete(r.users, id)
	tracing.AddSpanAttributes(span, attribute.String("operation.result", "success"))
	return nil
}
