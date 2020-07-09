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
	Add(hydration models.Hydration) (int64, error)
	// FetchDailyAmount gets summary of today's total drink amount.
	FetchDailyAmount(userName string) (int64, error)
	// Update updates hydration data.
	Update(hydration models.Hydration) error
	// Delete deletes hydration data.
	Delete(hydration models.Hydration) error
}
