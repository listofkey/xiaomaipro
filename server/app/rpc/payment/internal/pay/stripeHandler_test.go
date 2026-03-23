package pay

import (
	"context"
	"testing"
)

func TestDecodeStripeSessionSnapshotPrefersWebhookObject(t *testing.T) {
	raw := `{
		"id":"evt_123",
		"type":"checkout.session.completed",
		"data":{
			"object":{
				"id":"cs_test_123",
				"client_reference_id":"P123",
				"payment_intent":"pi_test_123",
				"payment_status":"paid",
				"status":"complete",
				"amount_total":18800
			}
		}
	}`

	snapshot, ok := decodeStripeSessionSnapshot(raw)
	if !ok {
		t.Fatalf("expected snapshot to decode")
	}
	if snapshot.ID != "cs_test_123" {
		t.Fatalf("expected session id, got %q", snapshot.ID)
	}
	if got := extractIDFromRawMessage(snapshot.PaymentIntent); got != "pi_test_123" {
		t.Fatalf("expected payment intent id, got %q", got)
	}
}

func TestResolvePaymentIntentIDKeepsSessionTradeNo(t *testing.T) {
	strategy := &StripeStrategy{}
	callbackData := `{
		"id":"evt_456",
		"type":"checkout.session.completed",
		"data":{
			"object":{
				"id":"cs_test_456",
				"payment_intent":"pi_test_456",
				"payment_status":"paid"
			}
		}
	}`

	paymentIntentID, err := strategy.resolvePaymentIntentID(context.Background(), "cs_test_456", callbackData)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if paymentIntentID != "pi_test_456" {
		t.Fatalf("expected payment intent id, got %q", paymentIntentID)
	}
}

func TestBuildRefundIdempotencyKeyStable(t *testing.T) {
	req := &RefundRequest{
		PaymentNo: "P123",
		Amount:    "188.00",
	}

	if got := buildRefundIdempotencyKey(req); got != "refund:P123:188.00" {
		t.Fatalf("unexpected idempotency key: %q", got)
	}
}
