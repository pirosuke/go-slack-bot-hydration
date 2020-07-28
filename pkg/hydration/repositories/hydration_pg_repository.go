package repositories

import (
	"context"
	"strconv"
	"strings"
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

// FetchWeeklyUsers returns user list with hydration record during this week.
func (repo *HydrationPgRepository) FetchWeeklyUsers() ([]string, error) {
	var userList []string

	rows, err := repo.conn.Query(context.Background(), "select username from hydrations where modified >= now()::date - interval '7 days' group by username")
	if err != nil {
		return userList, err
	}

	for rows.Next() {
		var userName string
		err := rows.Scan(&userName)
		if err != nil {
			return userList, err
		}
		userList = append(userList, userName)
	}

	return userList, nil
}

// FetchWeeklySummary returns summary of weekly hydration.
func (repo *HydrationPgRepository) FetchWeeklySummary(userName string) ([]models.DailyHydrationSummary, error) {
	var resultList []models.DailyHydrationSummary

	sql := []string{
		"select ",
		"extract(day from modified)::text as day ",
		",sum(amount) as total_amount ",
		"from hydrations ",
		"where username = $1 ",
		"and modified >= now()::date - interval '7 days' ",
		"group by extract(day from modified) ",
		"order by extract(day from modified)",
	}

	rows, err := repo.conn.Query(context.Background(), strings.Join(sql, " "), userName)
	if err != nil {
		return resultList, err
	}

	for rows.Next() {
		var day string
		var totalAmount int64
		err := rows.Scan(&day, &totalAmount)
		if err != nil {
			return resultList, err
		}
		resultList = append(resultList, models.DailyHydrationSummary{
			Day:         day,
			TotalAmount: totalAmount,
		})
	}

	return resultList, nil
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
