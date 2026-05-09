package webhook

// Webhook represents a registered webhook receiver.
type Webhook struct {
	ID  string `json:"webhook_id"`
	URL string `json:"url"`
}
