package repositories

import (
	"context"
	"strconv"
	"time"

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
func (repo *HydrationPgRepository) Add(hydration models.Hydration) (int64, error) {
	var hydrationID int64
	err := repo.conn.QueryRow(context.Background(),
		"insert into hydrations(username, drink, amount, modified) values($1, $2, $3, $4) returning id",
		hydration.Username,
		hydration.Drink,
		hydration.Amount,
		hydration.Modified,
	).Scan(&hydrationID)

	return hydrationID, err
}

// FetchOne fetches one hydration data.
func (repo *HydrationPgRepository) FetchOne(hydrationID int64) (models.Hydration, error) {
	var hydration models.Hydration
	var userName string
	var drink string
	var amount int64
	var modified time.Time
	err := repo.conn.QueryRow(context.Background(), "select username, drink, amount, modified from hydrations where id = $1",
		hydrationID,
	).Scan(
		&userName,
		&drink,
		&amount,
		&modified,
	)

	if err != nil {
		return hydration, err
	}

	hydration = models.Hydration{
		ID:       hydrationID,
		Username: userName,
		Drink:    drink,
		Amount:   amount,
		Modified: modified,
	}

	return hydration, nil
}

// FetchDailyAmount gets summary of today's total drink amount.
func (repo *HydrationPgRepository) FetchDailyAmount(userName string) (int64, error) {
	var totalAmount int64
	err := repo.conn.QueryRow(context.Background(), "select sum(amount) from hydrations where username = $1 and modified::date = now()::date", userName).Scan(&totalAmount)

	return totalAmount, err
}

// Update updates hydration data.
func (repo *HydrationPgRepository) Update(hydration models.Hydration) error {
	_, err := repo.conn.Exec(context.Background(), "update hydrations set drink = $1, amount = $2, modified = $3 where id = $4 and username = $5",
		hydration.Drink,
		hydration.Amount,
		hydration.Modified,
		hydration.ID,
		hydration.Username,
	)
	return err
}

// Delete deletes hydration data.
func (repo *HydrationPgRepository) Delete(hydration models.Hydration) error {
	_, err := repo.conn.Exec(context.Background(), "delete from hydrations where id = $1 and username = $2",
		hydration.ID,
		hydration.Username,
	)
	return err
}
