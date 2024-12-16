package gopay

import "fmt"

type (
	TransactionType string
	NetworkType     string
	NetworkMode     string
	Currency        string
	PaymentStatus   string
	PaymentType     string
)

const (
	FIAT   PaymentType = "FIAT"
	CRYPTO PaymentType = "CRYPTO"

	DEPOSIT TransactionType = "DEPOSIT"
	PAYOUT  TransactionType = "PAYOUT"

	EVM     NetworkType = "EVM"
	CARDANO NetworkType = "CARDANO"

	MAINNET NetworkMode = "MAINNET"
	TESTNET NetworkMode = "TESTNET"

	USD Currency = "USD"
	JPY Currency = "JPY"

	INITIATED       PaymentStatus = "INITIATED"
	PENDING_DEPOSIT PaymentStatus = "PENDING_DEPOSIT"
	DEPOSITED       PaymentStatus = "DEPOSITED"
	ON_HOLD         PaymentStatus = "ON_HOLD"
	PAID_OUT        PaymentStatus = "PAID_OUT"
	CANCLED         PaymentStatus = "CANCELED"
	REFUNDED        PaymentStatus = "REFUNDED"
)

func scanEnum(value interface{}, target interface{}) error {
	switch v := value.(type) {
	case []byte:
		*target.(*string) = string(v)
	case string:
		*target.(*string) = v
	default:
		return fmt.Errorf("failed to scan type: %v", value)
	}
	return nil
}

func (c *TransactionType) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c))
}

func (c *NetworkType) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c))
}

func (c *NetworkMode) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c))
}

func (c *Currency) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c))
}

func (c *PaymentStatus) Scan(value interface{}) error {
	return scanEnum(value, (*string)(c))
}
