// Package di provides dependency injection for the application.
// It initializes the database connection, queries, and services.
package di

// import (
// 	"github.com/EfosaE/credora-backend/internal/db"
// 	"github.com/EfosaE/credora-backend/internal/db/sqlc"
// 	"github.com/EfosaE/credora-backend/internal/services"
// )

// type Container struct {
// 	DB             *db.DB
// 	Queries        *sqlc.Queries
// 	UserService    *services.SqlcUserService
// 	// EmailService   services.EmailService
// 	// PaymentService services.PaymentService
// }

// func NewContainer() (*Container, error) {
// 	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	queries := sqlc.New(db)
// 	ctx := context.Background()

// 	return &Container{
// 		DB:             db,
// 		Queries:        queries,
// 		UserService:    services.NewSqlcUserService(ctx, queries),
// 		// EmailService:   services.NewSendGridService(...),
// 		// PaymentService: services.NewPaystackService(...),
// 	}, nil
// }
