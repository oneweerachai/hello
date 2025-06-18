package repository

import (
	"errors"
	"sync"
	"user-api/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
}

// InMemoryUserRepository implements UserRepository using in-memory storage
type InMemoryUserRepository struct {
	users map[string]*models.User
	mutex sync.RWMutex
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*models.User),
		mutex: sync.RWMutex{},
	}
}

// Create adds a new user to the repository
func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user with same email already exists
	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			return errors.New("user with this email already exists")
		}
	}

	r.users[user.ID] = user
	return nil
}

// GetByID retrieves a user by ID
func (r *InMemoryUserRepository) GetByID(id string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *InMemoryUserRepository) GetByEmail(email string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// GetAll retrieves all users
func (r *InMemoryUserRepository) GetAll() ([]*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

// Update updates an existing user
func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	r.users[user.ID] = user
	return nil
}

// Delete removes a user from the repository
func (r *InMemoryUserRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[id]; !exists {
		return errors.New("user not found")
	}

	delete(r.users, id)
	return nil
}
