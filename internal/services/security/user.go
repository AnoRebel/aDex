package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// User represents a user in the aDex-UI system
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"`
	DisplayName  string    `json:"displayName"`
	Role         UserRole  `json:"role"`
	Permissions  []string  `json:"permissions"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	LoginCount   int       `json:"loginCount"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UserRole defines user access levels
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RolePowerUser UserRole = "power_user"
	RoleUser      UserRole = "user"
	RoleGuest     UserRole = "guest"
)

// UserStore manages user storage and retrieval
type UserStore struct {
	users      map[string]*User // indexed by ID
	byUsername map[string]*User // indexed by username
	byEmail    map[string]*User // indexed by email
	storePath  string
	mu         sync.RWMutex
}

// NewUserStore creates a new user storage system
func NewUserStore(storePath string) (*UserStore, error) {
	store := &UserStore{
		users:      make(map[string]*User),
		byUsername: make(map[string]*User),
		byEmail:    make(map[string]*User),
		storePath:  storePath,
	}

	// Load existing users from storage
	if err := store.load(); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load users: %w", err)
		}
	}

	return store, nil
}

// CreateUser creates a new user
func (s *UserStore) CreateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate user
	if err := s.validateUser(user); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Check for duplicate username
	if _, exists := s.byUsername[user.Username]; exists {
		return fmt.Errorf("username already exists: %s", user.Username)
	}

	// Check for duplicate email if provided
	if user.Email != "" {
		if _, exists := s.byEmail[user.Email]; exists {
			return fmt.Errorf("email already exists: %s", user.Email)
		}
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Store user
	s.users[user.ID] = user
	s.byUsername[user.Username] = user
	if user.Email != "" {
		s.byEmail[user.Email] = user
	}

	// Save to disk
	return s.save()
}

// GetUser retrieves a user by ID
func (s *UserStore) GetUser(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserStore) GetUserByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.byUsername[username]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserStore) GetUserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.byEmail[email]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", email)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserStore) UpdateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	existingUser, exists := s.users[user.ID]
	if !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	// If username changed, check for duplicates
	if existingUser.Username != user.Username {
		if _, exists := s.byUsername[user.Username]; exists {
			return fmt.Errorf("username already exists: %s", user.Username)
		}
		// Remove old username index
		delete(s.byUsername, existingUser.Username)
	}

	// If email changed, check for duplicates
	if existingUser.Email != user.Email {
		if user.Email != "" {
			if _, exists := s.byEmail[user.Email]; exists {
				return fmt.Errorf("email already exists: %s", user.Email)
			}
		}
		// Remove old email index
		if existingUser.Email != "" {
			delete(s.byEmail, existingUser.Email)
		}
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Update indexes
	s.users[user.ID] = user
	s.byUsername[user.Username] = user
	if user.Email != "" {
		s.byEmail[user.Email] = user
	}

	return s.save()
}

// DeleteUser deletes a user
func (s *UserStore) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return fmt.Errorf("user not found: %s", id)
	}

	// Remove from all indexes
	delete(s.users, id)
	delete(s.byUsername, user.Username)
	if user.Email != "" {
		delete(s.byEmail, user.Email)
	}

	return s.save()
}

// ListUsers returns all users
func (s *UserStore) ListUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users
}

// UserCount returns the number of users
func (s *UserStore) UserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.users)
}

// UpdateLastLogin updates the user's last login timestamp
func (s *UserStore) UpdateLastLogin(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.LoginCount++
	user.UpdatedAt = now

	return s.save()
}

// Private methods

// validateUser validates user data
func (s *UserStore) validateUser(user *User) error {
	if user.ID == "" {
		return fmt.Errorf("user ID is required")
	}

	if user.Username == "" {
		return fmt.Errorf("username is required")
	}

	if len(user.Username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}

	if len(user.Username) > 32 {
		return fmt.Errorf("username must be at most 32 characters")
	}

	// Validate email format if provided
	if user.Email != "" && !isValidEmail(user.Email) {
		return fmt.Errorf("invalid email format: %s", user.Email)
	}

	// Validate role
	validRoles := map[UserRole]bool{
		RoleAdmin:     true,
		RolePowerUser: true,
		RoleUser:      true,
		RoleGuest:     true,
	}
	if !validRoles[user.Role] {
		return fmt.Errorf("invalid user role: %s", user.Role)
	}

	return nil
}

// save persists users to disk
func (s *UserStore) save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(s.storePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create user store directory: %w", err)
	}

	// Convert users map to slice
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(s.storePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write user store: %w", err)
	}

	return nil
}

// load reads users from disk
func (s *UserStore) load() error {
	data, err := os.ReadFile(s.storePath)
	if err != nil {
		return err
	}

	var users []*User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("failed to unmarshal users: %w", err)
	}

	// Rebuild indexes
	for _, user := range users {
		s.users[user.ID] = user
		s.byUsername[user.Username] = user
		if user.Email != "" {
			s.byEmail[user.Email] = user
		}
	}

	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Simple regex for basic email validation
	// In production, use a more robust validation library
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false // Multiple @ signs
			}
			atIndex = i
		}
	}

	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false // @ at beginning or end
	}

	// Check for dot after @
	hasDotAfterAt := false
	for i := atIndex + 1; i < len(email); i++ {
		if email[i] == '.' {
			hasDotAfterAt = true
			break
		}
	}

	return hasDotAfterAt
}
