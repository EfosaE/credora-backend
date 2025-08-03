package seeder

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Seeder struct {
	Conn *pgx.Conn
	Ctx  context.Context
}

func NewSeeder(conn *pgx.Conn, ctx context.Context) *Seeder {
	gofakeit.Seed(0)
	return &Seeder{
		Conn: conn,
		Ctx:  ctx,
	}
}

func (s *Seeder) SeedUsersAndAccounts(count int) error {
	for range count {
		now := time.Now()
		userID := uuid.New()
		userEmail := gofakeit.Email()

		_, err := s.Conn.Exec(s.Ctx, `
			INSERT INTO users (
				id, full_name, email, phone_number, password, is_verified,
				created_at, updated_at, nin, expires_at, monnify_customer_ref
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		`,
			userID,
			gofakeit.Name(),
			toText(userEmail),
			gofakeit.Phone(),
			gofakeit.Password(true, true, true, true, false, 12),
			toBool(gofakeit.Bool()),
			toTimestamp(now),
			toTimestamp(now),
			gofakeit.Numerify("###########"),
			toTimestamp(now.AddDate(1, 0, 0)),
			toText("ref_"+gofakeit.UUID()),
		)
		if err != nil {
			log.Printf("Failed to insert user: %v", err)
			continue
		}

		accountID := uuid.New()
		accountNumber := gofakeit.Numerify("3#########")
		_, err = s.Conn.Exec(s.Ctx, `
			INSERT INTO accounts (
				id, user_id, account_number, account_type,
				balance, currency, created_at, updated_at,
				virtual_account_bank, monnify_customer_ref
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`,
			accountID,
			toUUID(userID),
			accountNumber,
			"savings",
			toNumeric(gofakeit.Price(1000, 100000)),
			"NGN",
			toTimestamp(now),
			toTimestamp(now),
			toText("Wema Bank"),
			toText("ref_"+gofakeit.UUID()),
		)
		if err != nil {
			log.Printf("Failed to insert account: %v", err)
			continue
		}

		log.Printf("Seeded user %s (%s)", userID.String(), userEmail)
	}

	return nil
}

// Helper functions
func toText(s string) pgtype.Text         { return pgtype.Text{String: s, Valid: true} }
func toBool(b bool) pgtype.Bool           { return pgtype.Bool{Bool: b, Valid: true} }
func toUUID(id uuid.UUID) pgtype.UUID     { return pgtype.UUID{Bytes: id, Valid: true} }
func toTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}
func toNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.2f", f))
	return n
}
