package gopay

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

type Transaction struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	PaymentID  uuid.UUID       `db:"payment_id" json:"-"`
	IdentityID uuid.UUID       `db:"identity_id" json:"identity_id"`
	TXID       string          `db:"tx_id" json:"tx_id"`
	Tag        string          `db:"tag" json:"tag"`
	Amount     float64         `db:"amount" json:"amount"`
	Fee        float64         `db:"fee" json:"fee"`
	Discount   float64         `db:"discount" json:"discount"`
	Type       TransactionType `db:"type" json:"type"`
	Meta       types.JSONText  `db:"meta" json:"meta"`
	CanceledAt *time.Time      `db:"canceled_at" json:"canceled_at"`
	VerfiedAt  *time.Time      `db:"verified_at" json:"verified_at"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
}
