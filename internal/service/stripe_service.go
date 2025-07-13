package service

import (
	"fmt"

	"github.com/reazonvan/LootBay/pkg/logger"
)

type StripeService interface {
	CreatePaymentIntent(amount float64, currency string, metadata map[string]string) (*StripePaymentIntent, error)
	CapturePaymentIntent(paymentIntentID string) error
	RefundPayment(paymentIntentID string, amount float64) error
	VerifyWebhook(payload []byte, signature string) (map[string]interface{}, error)
}

type StripePaymentIntent struct {
	ID           string `json:"id"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ClientSecret string `json:"client_secret"`
}

type stripeService struct {
	secretKey string
	logger    *logger.Logger
}

func NewStripeService(secretKey string, logger *logger.Logger) StripeService {
	return &stripeService{
		secretKey: secretKey,
		logger:    logger,
	}
}

func (s *stripeService) CreatePaymentIntent(amount float64, currency string, metadata map[string]string) (*StripePaymentIntent, error) {
	// Конвертация в центы (Stripe работает в центах)
	amountInCents := int64(amount * 100)

	// Здесь будет реальная интеграция с Stripe API
	// Пока возвращаем mock данные
	paymentIntent := &StripePaymentIntent{
		ID:           fmt.Sprintf("pi_mock_%d", amountInCents),
		Amount:       amountInCents,
		Currency:     currency,
		Status:       "requires_payment_method",
		ClientSecret: fmt.Sprintf("pi_mock_%d_secret", amountInCents),
	}

	s.logger.Info("Stripe payment intent created", "id", paymentIntent.ID, "amount", amountInCents)
	return paymentIntent, nil
}

func (s *stripeService) CapturePaymentIntent(paymentIntentID string) error {
	// Здесь будет реальная логика захвата платежа через Stripe API
	s.logger.Info("Stripe payment intent captured", "id", paymentIntentID)
	return nil
}

func (s *stripeService) RefundPayment(paymentIntentID string, amount float64) error {
	amountInCents := int64(amount * 100)

	// Здесь будет реальная логика возврата через Stripe API
	s.logger.Info("Stripe payment refunded", "id", paymentIntentID, "amount", amountInCents)
	return nil
}

func (s *stripeService) VerifyWebhook(payload []byte, signature string) (map[string]interface{}, error) {
	// Здесь будет реальная проверка webhook через Stripe API
	s.logger.Info("Stripe webhook verified", "signature", signature)

	// Mock данные для webhook
	return map[string]interface{}{
		"id":   "evt_mock_123",
		"type": "payment_intent.succeeded",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":     "pi_mock_123",
				"status": "succeeded",
			},
		},
	}, nil
}
