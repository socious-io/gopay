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
	TXID        string      `json:"tx_id"`
	TotalAmount int64       `json:"total_amount"`
	Currency    string      `json:"currency"`
	Meta        interface{} `json:"meta"`
	Date        time.Time   `json:"date"`
	Confirmed   bool        `json:"confirmed"`
}

type FiatParams struct {
	ServiceName string
	Customer    string
	Description string
	Amount      float64
	Currency    Currency
	Transfer    *Transfer
}

func (fiats Fiats) Pay(params FiatParams) (*FiatTransactionInfo, error) {
	for _, f := range fiats {
		if params.ServiceName != f.Name {
			continue
		}
		switch f.Service {
		// TODO: add new fiat services here
		default:
			return f.StripePay(params)
		}
	}
	return nil, fmt.Errorf("service %s could not found", params.ServiceName)
}

func (f Fiat) StripePay(params FiatParams) (*FiatTransactionInfo, error) {
	// Setup Key
	stripe.Key = f.ApiKey

	list := paymentmethod.List(&stripe.PaymentMethodListParams{
		Customer: stripe.String(params.Customer),
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
		return nil, fmt.Errorf("card method %s could not be found", params.Customer)
	}

	intentParams := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(stripeAmount(params.Amount, params.Currency)),
		Currency:      stripe.String(string(params.Currency)),
		Customer:      stripe.String(params.Customer),
		PaymentMethod: stripe.String(method.ID),
		Description:   stripe.String(params.Description),
	}

	if params.Transfer != nil {
		intentParams.ConfirmationMethod = stripe.String("automatic")
		intentParams.ApplicationFeeAmount = stripe.Int64(int64(params.Amount - params.Transfer.Amount))
		intentParams.OnBehalfOf = stripe.String(params.Transfer.Destination)
		intentParams.TransferData = &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(params.Transfer.Destination),
		}
	}

	result, err := paymentintent.New(intentParams)

	if err != nil {
		return nil, err
	}
	info := &FiatTransactionInfo{
		TXID:        result.ID,
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
