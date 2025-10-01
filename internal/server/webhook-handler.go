package server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/square/square-go-sdk"
	squareclient "github.com/square/square-go-sdk/client"
)

// Receive Square webhook events
// (POST /webhooks/square)
func (chqs *CircleHQService) SquareRequest(w http.ResponseWriter, r *http.Request) {
	chqs.logger.InfoContext(r.Context(), "SquareRequest received", "method", r.Method, "url", r.URL.String())
	mu.Lock()
	defer mu.Unlock()

	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// reqSignature := r.Header.Get("X-Square-HmacSha256-Signature")
	// if !isFromSquare(chqs.signatureKey, reqSignature, body) {
	// 	// Signature is invalid. Return 403 Forbidden.
	// 	w.WriteHeader(http.StatusForbidden)
	// 	fmt.Fprintln(w, "Invalid signature")
	// 	chqs.logger.WarnContext(ctx, "Signature is valid")
	// 	return
	// }

	if err := chqs.core.HandleInventoryUpdateWebhook(ctx, body); err != nil {
		chqs.logger.ErrorContext(ctx, "Failed to handle inventory update webhook", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// The URL where event notifications are sent.
// The signature key defined for the subscription.
const (
	NOTIFICATION_URL = "http://square-webhook.mikuta0407.net/webhooks/square"
)

// Generate a signature from the notification url, signature key,
// and request body and compare it to the Square signature header.
func isFromSquare(signatureKey string, reqSignature string, body []byte) bool {
	client := squareclient.NewClient()

	fmt.Println(signatureKey, reqSignature)

	err := client.Webhooks.VerifySignature(
		context.TODO(),
		&square.VerifySignatureRequest{
			RequestBody:     string(body),
			SignatureHeader: reqSignature,
			SignatureKey:    signatureKey,
			NotificationURL: NOTIFICATION_URL,
		},
	)
	fmt.Println(err)
	return err == nil
}
