package models

import "time"

// Hydration describes hydration data.
type Hydration struct {
	Username string
	Drink    string
	Amount   int64
	Modified time.Time
}
