package gopay

import (
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
