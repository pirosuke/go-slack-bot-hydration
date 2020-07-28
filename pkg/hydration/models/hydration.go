package models

import "time"

type (
	// Hydration describes hydration data.
	Hydration struct {
		ID       int64
		Username string
		Drink    string
		Amount   int64
		Modified time.Time
	}

	// DailyHydrationSummary describes daily total amount of drinks
	DailyHydrationSummary struct {
		Day         string
		TotalAmount int64
	}
)
