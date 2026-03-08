package fakes

import (
	"context"
	"errors"
	"sync"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/google/uuid"
)

type MockUserRepo struct {
	mu sync.Mutex

	// Function overrides
	CreateFunc                         func(ctx context.Context, req *user.CreateUserRequest) (*user.User, error)
	GetByIDFunc                        func(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetByEmailFunc                     func(ctx context.Context, email string) (*user.User, error)
	UpdatePasswordFunc                 func(ctx context.Context, id uuid.UUID, hashedPassword string) error
	GetUserAccountsByAccountNumberFunc func(ctx context.Context, accountNumber string) (*user.User, error)
	GetUserAccountsByEmailFunc         func(ctx context.Context, email string) (*user.User, error)
	VerifyUserFunc                     func(ctx context.Context, id uuid.UUID) error

	// Call tracking
	GetUserAccountsByAccountNumberCalled bool
	GetUserAccountsByEmailCalled         bool
	VerifyUserCalled                     bool
	CreateCalled                         bool
	GetByIDCalled                        bool
	GetByEmailCalled                     bool
	UpdatePasswordCalled                 bool

	Users map[uuid.UUID]*user.User
}

func (m *MockUserRepo) Create(
	ctx context.Context,
	req *user.CreateUserRequest,
) (*user.User, error) {

	m.mu.Lock()
	m.CreateCalled = true
	defer m.mu.Unlock()

	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}

	if m.Users == nil {
		m.Users = make(map[uuid.UUID]*user.User)
	}

	u := &user.User{
		ID:    uuid.New(),
		Email: req.Email,
	}

	m.Users[u.ID] = u

	return u, nil
}

func (m *MockUserRepo) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*user.User, error) {

	m.mu.Lock()
	m.GetByIDCalled = true
	defer m.mu.Unlock()

	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	if m.Users == nil {
		return nil, errors.New("no users in mock")
	}

	u, ok := m.Users[id]
	if !ok {
		return nil, errors.New("user not found")
	}

	return u, nil
}

func (m *MockUserRepo) GetByEmail(
	ctx context.Context,
	email string,
) (*user.User, error) {

	m.mu.Lock()
	m.GetByEmailCalled = true
	defer m.mu.Unlock()

	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}

	if m.Users == nil {
		return nil, errors.New("no users in mock")
	}

	for _, u := range m.Users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, errors.New("user not found")
}

func (m *MockUserRepo) UpdatePassword(
	ctx context.Context,
	id uuid.UUID,
	hashedPassword string,
) error {

	m.mu.Lock()
	m.UpdatePasswordCalled = true
	defer m.mu.Unlock()

	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, id, hashedPassword)
	}

	u, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}

	u.Password = hashedPassword
	return nil
}

func (m *MockUserRepo) GetUserAccountsByAccountNumber(
	ctx context.Context,
	accountNumber string,
) (*user.User, error) {

	m.mu.Lock()
	m.GetUserAccountsByAccountNumberCalled = true
	defer m.mu.Unlock()

	if m.GetUserAccountsByAccountNumberFunc != nil {
		return m.GetUserAccountsByAccountNumberFunc(ctx, accountNumber)
	}

	if m.Users == nil {
		return nil, errors.New("no users in mock")
	}

	for _, u := range m.Users {
		for _, acc := range u.Accounts {
			if acc.AccountNumber == accountNumber {
				return u, nil
			}
		}
	}

	return nil, errors.New("user not found")
}

func (m *MockUserRepo) GetUserAccountsByEmail(
	ctx context.Context,
	email string,
) (*user.User, error) {

	m.mu.Lock()
	m.GetUserAccountsByEmailCalled = true
	defer m.mu.Unlock()

	if m.GetUserAccountsByEmailFunc != nil {
		return m.GetUserAccountsByEmailFunc(ctx, email)
	}

	if m.Users == nil {
		return nil, errors.New("no users in mock")
	}

	for _, u := range m.Users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, errors.New("user not found")
}

func (m *MockUserRepo) VerifyUser(
	ctx context.Context,
	id uuid.UUID,
) error {

	m.mu.Lock()
	m.VerifyUserCalled = true
	defer m.mu.Unlock()

	if m.VerifyUserFunc != nil {
		return m.VerifyUserFunc(ctx, id)
	}

	u, ok := m.Users[id]
	if !ok {
		return errors.New("user verification failed:user not found")
	}

	u.IsVerified = true
	return nil
}
