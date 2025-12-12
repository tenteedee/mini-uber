package stripe

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/tenteedee/mini-uber/services/payment-service/internal/domain"
	"github.com/tenteedee/mini-uber/services/payment-service/pkg/types"
)

type StripeClient struct {
	config *types.PaymentConfig
}

func NewStripeClient(cfg *types.PaymentConfig) domain.PaymentProcessor {
	stripe.Key = cfg.StripeSecretKey
	return &StripeClient{
		config: cfg,
	}
}

func (s *StripeClient) CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(s.config.SuccessURL),
		CancelURL:  stripe.String(s.config.CancelURL),
		Metadata:   metadata,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Ride Payment"),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}
	result, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create a payment session on Stripe: %v", err)
	}

	return result.ID, nil
}
