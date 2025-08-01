package utils

import (
	crypto "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var maxFee = decimal.NewFromFloat(100)    // ₦100 cap
var feeRate = decimal.NewFromFloat(0.005) // 0.5%

func CalculateSettlement(amount decimal.Decimal) string {
	fee := amount.Mul(feeRate)

	// Cap fee at ₦100
	if fee.GreaterThan(maxFee) {
		fee = maxFee
	}

	return amount.Sub(fee).StringFixed(2)
}

// GenerateTransactionReference creates a secure transaction reference in the format:
// CRD-<userID>-<YYYYMMDD>-<secureRandom6Digits>
func GenerateTransactionReference(userID uuid.UUID) string {
	now := time.Now().Format("20060102") // e.g. 20250801

	// Generate secure 6-digit random number
	n, err := crypto.Int(crypto.Reader, big.NewInt(1000000))
	if err != nil {
		panic(fmt.Errorf("failed to generate secure random number: %w", err))
	}

	shortID := userID.String()[:8]

	return fmt.Sprintf("CRD-%s-%s-%06d", shortID, now, n.Int64())
}

func GenerateMonnifyReference(channelCode int) string {
	now := time.Now().Format("20060102150405") // e.g. 20250801124533
	sequence := rand.Intn(999999)              // e.g. 12345
	return fmt.Sprintf("MNFY|%02d|%s|%06d", channelCode, now, sequence)
}
