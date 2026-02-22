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
	// Different data every run
	gofakeit.Seed(time.Now().UnixNano())

	return &Seeder{
		Conn: conn,
		Ctx:  ctx,
	}
}

func (s *Seeder) SeedUsersAndAccounts(count int) error {
	start := time.Now()

	tx, err := s.Conn.Begin(s.Ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(s.Ctx)

	log.Printf("Starting seeding of %d users...", count)

	for i := 0; i < count; i++ {
		now := time.Now()
		userID := uuid.New()
		userEmail := gofakeit.Email()
		userName := gofakeit.Name()

		// --- USER ---
		_, err := tx.Exec(s.Ctx, `
			INSERT INTO users (
				id, full_name, email, phone_number, password, is_verified,
				created_at, updated_at, nin, expires_at, monnify_customer_ref
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		`,
			userID,
			userName,
			toText(userEmail),
			gofakeit.Phone(),
			gofakeit.Password(true, true, true, true, false, 12),
			toBool(true),
			toTimestamp(now),
			toTimestamp(now),
			gofakeit.Numerify("###########"),
			toTimestamp(now.AddDate(1, 0, 0)),
			toText("ref_"+gofakeit.UUID()),
		)
		if err != nil {
			return fmt.Errorf("user insert failed at %d: %w", i, err)
		}

		// --- ACCOUNT (STRICTLY 1:1) ---
		_, err = tx.Exec(s.Ctx, `
			INSERT INTO accounts (
				id, user_id,username, account_number, account_type,
				balance, currency, created_at, updated_at,
				virtual_account_bank, monnify_customer_ref
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10, $11)
		`,
			uuid.New(),
			toUUID(userID),
			userName,
			gofakeit.Numerify("3#########"),
			randomAccountType(),
			toNumeric(randomTieredBalance()),
			"NGN",
			toTimestamp(now),
			toTimestamp(now),
			toText("Wema Bank"),
			toText("ref_"+gofakeit.UUID()),
		)
		if err != nil {
			return fmt.Errorf("account insert failed at %d: %w", i, err)
		}

		// Progress every 100
		if (i+1)%100 == 0 {
			log.Printf("Progress: %d/%d users seeded...", i+1, count)
		}
	}

	if err := tx.Commit(s.Ctx); err != nil {
		return err
	}

	log.Printf("Done. Seeded %d users in %s", count, time.Since(start))
	return nil
}

// func (s *Seeder) SeedUsersAndAccounts(count int) error {
// 	tx, err := s.Conn.Begin(s.Ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback(s.Ctx)

// 	for i := 0; i < count; i++ {
// 		now := time.Now()
// 		userID := uuid.New()
// 		userEmail := gofakeit.Email()

// 		// Insert user
// 		_, err := tx.Exec(s.Ctx, `
// 			INSERT INTO users (
// 				id, full_name, email, phone_number, password, is_verified,
// 				created_at, updated_at, nin, expires_at, monnify_customer_ref
// 			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
// 		`,
// 			userID,
// 			gofakeit.Name(),
// 			toText(userEmail),
// 			gofakeit.Phone(),
// 			gofakeit.Password(true, true, true, true, false, 12),
// 			toBool(gofakeit.Bool()),
// 			toTimestamp(now),
// 			toTimestamp(now),
// 			gofakeit.Numerify("###########"),
// 			toTimestamp(now.AddDate(1, 0, 0)),
// 			toText("ref_"+gofakeit.UUID()),
// 		)
// 		if err != nil {
// 			log.Printf("User insert failed: %v", err)
// 			continue
// 		}

// 		// Each user can have 1–3 accounts
// 		accountCount := gofakeit.IntRange(1, 3)

// 		for j := 0; j < accountCount; j++ {
// 			accountID := uuid.New()

// 			_, err = tx.Exec(s.Ctx, `
// 				INSERT INTO accounts (
// 					id, user_id, account_number, account_type,
// 					balance, currency, created_at, updated_at,
// 					virtual_account_bank, monnify_customer_ref
// 				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
// 			`,
// 				accountID,
// 				toUUID(userID),
// 				gofakeit.Numerify("3#########"),
// 				randomAccountType(),
// 				toNumeric(randomTieredBalance()),
// 				"NGN",
// 				toTimestamp(now),
// 				toTimestamp(now),
// 				toText("Wema Bank"),
// 				toText("ref_"+gofakeit.UUID()),
// 			)

// 			if err != nil {
// 				log.Printf("Account insert failed: %v", err)
// 			}
// 		}

// 		if i%1000 == 0 {
// 			log.Printf("Seeded %d users...", i)
// 		}
// 	}

// 	return tx.Commit(s.Ctx)
// }

func (s *Seeder) SeedHotAccounts() error {
	now := time.Now()

	hotAccounts := []struct {
		accountType string
		balance     float64
	}{
		{"SETTLEMENT", 50_000_000},
		{"SALARY_POOL", 20_000_000},
		{"MERCHANT_WALLET", 15_000_000},
		{"SYSTEM_RESERVE", 100_000_000},
	}

	tx, err := s.Conn.Begin(s.Ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(s.Ctx)

	for _, acc := range hotAccounts {
		_, err := tx.Exec(s.Ctx, `
			INSERT INTO accounts (
				id,username, account_number, account_type,
				balance, currency, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7, $8)
		`,
			uuid.New(),
			gofakeit.Name(),
			gofakeit.Numerify("9#########"),
			acc.accountType,
			toNumeric(acc.balance),
			"NGN",
			toTimestamp(now),
			toTimestamp(now),
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit(s.Ctx)
}

// Helper functions
func toText(s string) pgtype.Text     { return pgtype.Text{String: s, Valid: true} }
func toBool(b bool) pgtype.Bool       { return pgtype.Bool{Bool: b, Valid: true} }
func toUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }
func toTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}
func toNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.2f", f))
	return n
}
func randomTieredBalance() float64 {
	r := gofakeit.IntRange(1, 100)

	switch {
	case r <= 60:
		return gofakeit.Price(1000, 10000)
	case r <= 90:
		return gofakeit.Price(50000, 200000)
	default:
		return gofakeit.Price(1_000_000, 5_000_000)
	}
}
func randomAccountType() string {
	types := []string{"savings", "current", "business"}
	return types[gofakeit.Number(0, len(types)-1)]
}
