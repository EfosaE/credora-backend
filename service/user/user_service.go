package usersvc

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/logger"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/EfosaE/credora-backend/service"
)

type UserService struct {
	userRepo      user.UserRepository
	logger        zerolog.Logger
	eventBus      event.EventBus
	monnifySvc    *service.MonnifyService
	usrDeviceRepo user.DeviceRepository
	queue         queues.Queue
}

func NewUserService(
	userRepo user.UserRepository,
	log zerolog.Logger,
	eventBus event.EventBus,
	monnifySvc *service.MonnifyService,
	queue queues.Queue,
	device user.DeviceRepository,
) *UserService {
	serviceLogger := log.With().
		Str("service", "user-service").
		Logger()

	return &UserService{
		userRepo:      userRepo,
		logger:        serviceLogger,
		eventBus:      eventBus,
		monnifySvc:    monnifySvc,
		usrDeviceRepo: device,
		queue:         queue,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {

	log := logger.FromCtx(ctx, s.logger).With().
		Str("email", req.Email).
		Str("user_name", req.Name).
		Logger()

	log.Info().Msg("user creation initiated")

	hashedPassword, err := authsvc.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash password")
		return nil, err
	}
	req.Password = hashedPassword

	result, err := s.userRepo.Create(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("failed to create user in repository")
		return nil, err
	}

	log = log.With().Str("user_id", result.ID.String()).Logger()
	log.Info().Msg("user persisted successfully")

	monnifyCustResp, err := s.CreateVirtualAccount(ctx, req, result.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("failed to create monnify virtual account")
		return nil, err
	}

	accountRef := monnifyCustResp.ResponseBody.AccountReference
	accounts := monnifyCustResp.ResponseBody.Accounts

	log = log.With().
		Str("account_reference", accountRef).
		Str("reservation_ref", monnifyCustResp.ResponseBody.ReservationReference).
		Logger()

	log.Info().Msg("virtual account created successfully")

	if err := s.SendAccountNumberEmailAsync(result.Email, accounts); err != nil {
		log.Error().Err(err).Msg("failed to enqueue account number email")
	} else {
		log.Info().Msg("account number email enqueued successfully")
	}

	evt := event.UserCreatedEvent{
		UserID:   result.ID,
		Accounts: accounts,
		Name:     result.FullName,
		Email:    result.Email,
	}

	payload, err := utils.StructToMap(evt)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert user created event to payload map")
		return nil, fmt.Errorf("failed to convert typed struct to map: %w", err)
	}

	if err := s.eventBus.Publish(ctx, event.StreamUserEvents, event.EventUserCreated, payload); err != nil {
		log.Error().Err(err).Str("event_type", event.EventUserCreated).Msg("failed to publish user created event")
	} else {
		log.Info().Str("event_type", event.EventUserCreated).Msg("user created event published successfully")
	}

	log.Info().Msg("user creation flow completed successfully")

	return &user.CreateUserResponse{
		ID:                   result.ID,
		Name:                 req.Name,
		Email:                req.Email,
		AccountReference:     accountRef,
		ReservationReference: monnifyCustResp.ResponseBody.ReservationReference,
		Status:               monnifyCustResp.ResponseBody.Status,
		CreatedAt:            result.CreatedAt,
	}, nil
}

func (s *UserService) RegisterDeviceTokenToUserID(ctx context.Context, userID uuid.UUID, token, platform string) (*user.DeviceToken, error) {

	log := logger.FromCtx(ctx, s.logger).With().
		Str("user_id", userID.String()).
		Str("platform", platform).
		Logger()

	log.Info().Msg("registering device token for user")

	dt, err := s.usrDeviceRepo.Create(ctx, userID, token, platform)
	if err != nil {
		log.Error().Err(err).Msg("failed to register device token")
		return nil, err
	}

	log.Info().Msg("device token registered successfully")
	return dt, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {

	log := logger.FromCtx(ctx, s.logger).With().
		Str("email", email).
		Logger()

	log.Info().Msg("fetching user by email")

	usr, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch user by email")
		return nil, err
	}

	log.Info().Str("user_id", usr.ID.String()).Msg("user fetched successfully")
	return usr, nil
}
