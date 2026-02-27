package fakes

import (
	"context"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/google/uuid"
)

type MockDeviceTokenRepository struct {
	CreateFunc         func(ctx context.Context, userID uuid.UUID, token, platform string) (*user.DeviceToken, error)
	GetByTokenFunc     func(ctx context.Context, token string) (*user.DeviceToken, error)
	GetByUserIDFunc    func(ctx context.Context, userID uuid.UUID) ([]*user.DeviceToken, error)
	UpdateFunc         func(ctx context.Context, id int64, token, platform string) (*user.DeviceToken, error)
	DeleteFunc         func(ctx context.Context, id int64) error
	DeleteByUserIDFunc func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockDeviceTokenRepository) Create(ctx context.Context, userID uuid.UUID, token, platform string) (*user.DeviceToken, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, token, platform)
	}
	return nil, nil
}

// func (m *MockDeviceTokenRepository) GetByToken(ctx context.Context, token string) (*user.DeviceToken, error) {
// 	if m.GetByTokenFunc != nil {
// 		return m.GetByTokenFunc(ctx, token)
// 	}
// 	return nil, nil
// }

func (m *MockDeviceTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*user.DeviceToken, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockDeviceTokenRepository) Update(ctx context.Context, id int64, token, platform string) (*user.DeviceToken, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, token, platform)
	}
	return nil, nil
}

func (m *MockDeviceTokenRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockDeviceTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteByUserIDFunc != nil {
		return m.DeleteByUserIDFunc(ctx, userID)
	}
	return nil
}

// package fakes

// import (
// 	"context"

// 	"github.com/EfosaE/credora-backend/domain/user"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/mock"
// )

// type MockDeviceTokenRepository struct {
// 	mock.Mock
// }

// func (m *MockDeviceTokenRepository) Create(ctx context.Context, userID uuid.UUID, token, platform string) (*user.DeviceToken, error) {
// 	args := m.Called(ctx, userID, token, platform)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*user.DeviceToken), args.Error(1)
// }

// func (m *MockDeviceTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*user.DeviceToken, error) {
// 	args := m.Called(ctx, userID)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).([]*user.DeviceToken), args.Error(1)
// }

// func (m *MockDeviceTokenRepository) Update(ctx context.Context, id int64, token, platform string) (*user.DeviceToken, error) {
// 	args := m.Called(ctx, id, token, platform)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*user.DeviceToken), args.Error(1)
// }

// func (m *MockDeviceTokenRepository) Delete(ctx context.Context, id int64) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }

// func (m *MockDeviceTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
// 	args := m.Called(ctx, userID)
// 	return args.Error(0)
// }
