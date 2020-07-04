package interfaces

import (
	"github.com/pirosuke/slack-bot-hydration/internal/database"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/models"
)

// HydrationRepository is interface for hydration repository.
type HydrationRepository interface {
	// Connect connects to database.
	Connect(config database.DbConfig) error
	// Close closes connection to database.
	Close()
	// Add inserts hydration data.
	Add(hydration models.Hydration) error
	// FetchDailyAmount gets summary of today's total drink amount.
	FetchDailyAmount(userName string) int64
}
