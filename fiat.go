package gopay

import (
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/paymentmethod"
)

type Fiats []Fiat

type Fiat struct {
	Name    string
	ApiKey  string
	Service FiatService
}

type Transfer struct {
	Amount      float64
	Destination string
}

type FiatTransactionInfo struct {
	TxID        string      `json:"txid"`
	TotalAmount int64       `json:"total_amount"`
	Currency    string      `json:"currency"`
	Meta        interface{} `json:"meta"`
	Date        time.Time   `json:"date"`
	Confirmed   bool        `json:"confirmed"`
}

func (fiats Fiats) Pay(serviceName, customer, description string, amount float64, currency Currency, transfer *Transfer) (*FiatTransactionInfo, error) {
	for _, f := range fiats {
		if serviceName != f.Name {
			continue
		}
		switch f.Service {
		// TODO: add new fiat services here
		default:
			return f.StripePay(customer, description, amount, currency, transfer)
		}
	}
	return nil, fmt.Errorf("service %s could not found", serviceName)
}

func (f Fiat) StripePay(customer, description string, amount float64, currency Currency, transfer *Transfer) (*FiatTransactionInfo, error) {
	// Setup Key
	stripe.Key = f.ApiKey

	list := paymentmethod.List(&stripe.PaymentMethodListParams{
		Customer: stripe.String(customer),
		Type:     stripe.String("card"),
	})
	var method *stripe.PaymentMethod
	for list.Next() {
		method = list.PaymentMethod()
	}
	if err := list.Err(); err != nil {
		return nil, err
	}
	if method == nil {
		return nil, fmt.Errorf("card method %s could not be found", customer)
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(stripeAmount(amount, currency)),
		Currency:      stripe.String(string(currency)),
		Customer:      stripe.String(customer),
		PaymentMethod: stripe.String(method.ID),
		Description:   stripe.String(description),
	}

	if transfer != nil {
		params.ConfirmationMethod = stripe.String("automatic")
		params.ApplicationFeeAmount = stripe.Int64(int64(amount - transfer.Amount))
		params.OnBehalfOf = stripe.String(transfer.Destination)
		params.TransferData = &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(transfer.Destination),
		}
	}

	result, err := paymentintent.New(params)

	if err != nil {
		return nil, err
	}
	info := &FiatTransactionInfo{
		TxID:        result.ID,
		TotalAmount: result.Amount,
		Date:        time.Now(),
		Currency:    string(result.Currency),
		Meta:        result,
	}

	if _, err := paymentintent.Confirm(
		result.ID,
		&stripe.PaymentIntentConfirmParams{
			PaymentMethod: stripe.String(method.ID),
		},
	); err != nil {
		return info, err
	}
	info.Confirmed = true
	return info, nil
}

func stripeAmount(amount float64, currency Currency) int64 {
	switch currency {
	case USD:
		return int64(amount * 100)
	case JPY:
		return int64(amount)
	default:
		return 0
	}
}
