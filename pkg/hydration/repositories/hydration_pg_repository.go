package repositories

import (
	"context"
	"fmt"
	"strconv"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pirosuke/slack-bot-hydration/internal/database"
	"github.com/pirosuke/slack-bot-hydration/pkg/hydration/models"
)

// HydrationPgRepository is gateway for PostgreSQL database repository.
type HydrationPgRepository struct {
	conn *pgx.Conn
}

// Connect connects to database.
func (repo *HydrationPgRepository) Connect(config database.DbConfig) error {
	var err error
	dbURL := "postgres://" + config.Connection.User + ":" + config.Connection.Password + "@" + config.Connection.Host + ":" + strconv.FormatInt(config.Connection.Port, 10) + "/" + config.Connection.Database
	repo.conn, err = pgx.Connect(context.Background(), dbURL)
	return err
}

// Close closes connection to database.
func (repo *HydrationPgRepository) Close() {
	repo.conn.Close(context.Background())
}

// Add inserts hydration data.
func (repo *HydrationPgRepository) Add(hydration models.Hydration) error {
	_, err := repo.conn.Exec(context.Background(),
		"insert into hydrations(username, drink, amount, modified) values($1, $2, $3, $4)",
		hydration.Username,
		hydration.Drink,
		hydration.Amount,
		hydration.Modified,
	)

	return err
}

// FetchDailyAmount gets summary of today's total drink amount.
func (repo *HydrationPgRepository) FetchDailyAmount(userName string) int64 {
	var totalAmount int64
	err := repo.conn.QueryRow(context.Background(), "select sum(amount) from hydrations where username = $1 and modified::date = now()::date", userName).Scan(&totalAmount)
	if err != nil {
		fmt.Println(err)
	}

	return totalAmount
}
