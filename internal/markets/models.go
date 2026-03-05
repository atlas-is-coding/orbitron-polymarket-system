package markets

import "time"

// AlertRule defines a one-shot price threshold alert for a market.
type AlertRule struct {
	ID          string
	ConditionID string
	TokenID     string  // YES token ID
	Direction   string  // "above" or "below"
	Threshold   float64
	CreatedAt   time.Time
	Triggered   bool
}
