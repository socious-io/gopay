package gopay

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

// Transaction represents a financial transaction related to a payment.
// It includes details about the transaction ID, amount, fees, discounts, and the associated payment and identity.
type Transaction struct {
	ID         uuid.UUID       `db:"id" json:"id"`                   // Unique transaction identifier
	PaymentID  uuid.UUID       `db:"payment_id" json:"-"`            // Associated payment ID (hidden in JSON)
	IdentityID uuid.UUID       `db:"identity_id" json:"identity_id"` // Associated identity ID
	TXID       string          `db:"tx_id" json:"tx_id"`             // Transaction ID (e.g., blockchain TX ID)
	Tag        string          `db:"tag" json:"tag"`                 // Tag associated with the transaction
	Amount     float64         `db:"amount" json:"amount"`           // Transaction amount
	Fee        float64         `db:"fee" json:"fee"`                 // Fee applied to the transaction
	Discount   float64         `db:"discount" json:"discount"`       // Discount applied to the transaction
	Status     *string         `db:"status" json:"status"`
	Type       TransactionType `db:"type" json:"type"`               // Type of the transaction (e.g., deposit, withdrawal)
	Meta       types.JSONText  `db:"meta" json:"meta"`               // Metadata associated with the transaction
	CanceledAt *time.Time      `db:"canceled_at" json:"canceled_at"` // Cancellation timestamp, if applicable
	VerfiedAt  *time.Time      `db:"verified_at" json:"verified_at"` // Verification timestamp
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`   // Transaction creation timestamp
}

// Table returns the table name for the Transaction struct, including a prefix if defined in config.
func (Transaction) Table() string {
	if config.Prefix == "" {
		return "transactions" // Default table name
	}
	return fmt.Sprintf("%s_transactions", config.Prefix) // Prefixed table name
}

// Create inserts a new transaction into the database, using the fields in the Transaction struct.
// It returns an error if the insert fails.
func (t *Transaction) Create() error {
	// SQL query to insert a new transaction
	query := `
		INSERT INTO %s (
			payment_id, identity_id, tx_id, tag, amount, fee, discount, type, meta
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING *
	`
	query = fmt.Sprintf(query, t.Table())

	// Execute the insert query and scan the result back into the struct
	return config.DB.QueryRowx(query, t.PaymentID, t.IdentityID, t.TXID, t.Tag, t.Amount, t.Fee, t.Discount, t.Type, t.Meta).
		StructScan(t)
}

// Verify updates the transaction's status to verified, setting the transaction ID, metadata, and verification timestamp.
// It returns an error if the update fails.
func (t *Transaction) Verify() error {
	// SQL query to update a transaction as verified
	query := `UPDATE %s SET tx_id=$2, meta=$3, verified_at=NOW() WHERE id=$1 RETURNING *`
	query = fmt.Sprintf(query, t.Table())

	// Execute the update query and scan the result back into the struct
	return config.DB.QueryRowx(query, t.ID, t.TXID, t.Meta).StructScan(t)
}

// Cancel sets the canceled timestamp for a transaction and updates its metadata.
// It returns an error if the cancel operation fails.
func (t *Transaction) Cancel() error {
	// SQL query to update a transaction as canceled
	query := `UPDATE %s SET meta=$2, canceled_at=NOW() WHERE id=$1 RETURNING *`
	query = fmt.Sprintf(query, t.Table())

	// Execute the update query and scan the result back into the struct
	return config.DB.QueryRowx(query, t.ID, t.Meta).StructScan(t)
}

func (t *Transaction) ActionRequired() error {
	// SQL query to update a transaction as verified
	query := `UPDATE %s SET tx_id=$2, meta=$3, status='ACTION_REQUIRED' WHERE id=$1 RETURNING *`
	query = fmt.Sprintf(query, t.Table())

	// Execute the update query and scan the result back into the struct
	return config.DB.QueryRowx(query, t.ID, t.TXID, t.Meta).StructScan(t)
}
