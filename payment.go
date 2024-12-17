package gopay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

type Payment struct {
	ID                 uuid.UUID      `db:"id" json:"id"`
	Tag                string         `db:"tag" json:"tag"`
	Description        string         `db:"description" json:"description"`
	UniqueRef          string         `db:"unique_ref" json:"unique_ref"`
	TotalAmount        float64        `db:"total_amount" json:"total_amount"`
	Currency           Currency       `db:"currency" json:"currency"`
	FiatServiceName    *string        `db:"fiat_service_name" json:"fiat_service_name"`
	CryptoCurrency     *string        `db:"crypto_currency" json:"crypto_currency"`
	CryptoCurrencyRate float64        `db:"crypto_currency_rate" json:"crypto_currency_rate"`
	Meta               types.JSONText `db:"meta" json:"meta,omitempty"`
	Status             PaymentStatus  `db:"status" json:"status"`
	Type               PaymentType    `db:"type" json:"type"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time      `db:"created_at" json:"updated_at"`
}

type PaymentIdentity struct {
	ID              uuid.UUID      `db:"id" json:"id"`
	PaymentID       uuid.UUID      `db:"payment_id" json:"payment_id"`
	IdentityID      uuid.UUID      `db:"identity_id" json:"identity_id"`
	RoleName        string         `db:"role_name" json:"role_name"`
	AllocatedAmount float64        `db:"allocated_amount" json:"allocated_amount"`
	Meta            types.JSONText `db:"meta" json:"meta,omitempty"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
}

func New(tag, desc, ref, string, currency Currency, meta interface{}) (*Payment, error) {
	// Convert meta to JSONB
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta: %w", err)
	}

	// Prepare the Payment struct for scanning
	payment := new(Payment)

	// SQL query with RETURNING *
	query := `
			INSERT INTO payments (tag, description, unique_ref, total_amount, currency, status, meta)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING *`

	// Execute query and scan the returned row into the struct
	if err := config.DB.QueryRowx(query, tag, desc, ref, 0.0, currency, INITIATED, metaJSON).
		StructScan(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (p *Payment) AddIdentity(id uuid.UUID, roleName string, amount float64, meta interface{}) (*PaymentIdentity, error) {

	// Convert meta to JSONB
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta: %w", err)
	}

	// Prepare PaymentIdentity struct for scanning
	identity := new(PaymentIdentity)

	// SQL query with RETURNING *
	query := `
		INSERT INTO payment_identities (payment_id, identity_id, role_name, allocated_amount, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING *`

	// Execute query and scan the returned row into the struct
	if err := config.DB.QueryRowx(query, p.ID, id, roleName, amount, metaJSON).
		StructScan(identity); err != nil {
		return nil, err
	}
	return identity, nil
}

func (p *Payment) SetToCryptoMode(address string, rate float64) error {

	// SQL query with RETURNING *
	query := `
		UPDATE payments
		SET crypto_currency = $1, crypto_currency_rate = $2, type = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING *`

	// Execute query and scan the returned row back into the Payment struct
	if err := config.DB.QueryRowx(query, address, rate, CRYPTO, p.ID).StructScan(p); err != nil {
		return fmt.Errorf("failed to set payment to crypto mode: %w", err)
	}

	return nil
}

func (p *Payment) SetToFiatMode(name string) error {
	// SQL query with RETURNING *
	query := `
		UPDATE payments
		SET fiat_service_name = $1, type = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING *`

	// Execute query and scan the updated row back into the Payment struct
	if err := config.DB.QueryRowx(query, name, FIAT, p.ID).
		StructScan(p); err != nil {
		return fmt.Errorf("failed to set payment to fiat mode: %w", err)
	}

	return nil
}
