package analytics

import (
	"database/sql"
	"time"
)

type AnalyticsRepository struct {
	DB *sql.DB
}

type AnalyticsResult struct {
	Date              string  `json:"date"`
	TotalTransactions int     `json:"total_transactions"`
	TotalProfit       int64   `json:"total_profit"`
	TotalRevenue      int64   `json:"total_revenue"`
	SuccessRate       float64 `json:"success_rate"`
}

type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{
		DB: db,
	}
}

// GetAnalyticsByDateRange - Mendapatkan analytics berdasarkan range tanggal
func (repo *AnalyticsRepository) GetAnalyticsByDateRange(startDate, endDate time.Time) ([]AnalyticsResult, error) {
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total_transactions,
			SUM(profit_amount) as total_profit,
			SUM(price) as total_revenue,
			ROUND(
				(COUNT(CASE WHEN status = 'success' THEN 1 END) * 100.0 / COUNT(*)), 2
			) as success_rate
		FROM transactions 
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at) DESC
	`

	rows, err := repo.DB.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AnalyticsResult
	for rows.Next() {
		var result AnalyticsResult
		err := rows.Scan(
			&result.Date,
			&result.TotalTransactions,
			&result.TotalProfit,
			&result.TotalRevenue,
			&result.SuccessRate,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetAnalyticsByDate - Mendapatkan analytics untuk tanggal tertentu
func (repo *AnalyticsRepository) GetAnalyticsByDate(date time.Time) (*AnalyticsResult, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total_transactions,
			COALESCE(SUM(profit_amount), 0) as total_profit,
			COALESCE(SUM(price), 0) as total_revenue,
			CASE 
				WHEN COUNT(*) = 0 THEN 0 
				ELSE ROUND((COUNT(CASE WHEN status = 'success' THEN 1 END) * 100.0 / COUNT(*)), 2)
			END as success_rate
		FROM transactions 
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
	`

	row := repo.DB.QueryRow(query, startOfDay, endOfDay)

	var result AnalyticsResult
	err := row.Scan(
		&result.Date,
		&result.TotalTransactions,
		&result.TotalProfit,
		&result.TotalRevenue,
		&result.SuccessRate,
	)

	if err == sql.ErrNoRows {
		// Return empty result if no transactions found
		return &AnalyticsResult{
			Date:              date.Format("2006-01-02"),
			TotalTransactions: 0,
			TotalProfit:       0,
			TotalRevenue:      0,
			SuccessRate:       0,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTodayAnalytics - Mendapatkan analytics hari ini
func (repo *AnalyticsRepository) GetTodayAnalytics() (*AnalyticsResult, error) {
	now := time.Now()
	return repo.GetAnalyticsByDate(now)
}

// GetMonthlyAnalytics - Mendapatkan analytics bulanan
func (repo *AnalyticsRepository) GetMonthlyAnalytics(year int, month time.Month) ([]AnalyticsResult, error) {
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	return repo.GetAnalyticsByDateRange(startDate, endDate)
}

// GetAnalyticsSummary - Mendapatkan ringkasan analytics untuk range tanggal
func (repo *AnalyticsRepository) GetAnalyticsSummary(startDate, endDate time.Time) (*AnalyticsResult, error) {
	query := `
		SELECT 
			CONCAT($1::date, ' - ', $2::date) as date_range,
			COUNT(*) as total_transactions,
			COALESCE(SUM(profit_amount), 0) as total_profit,
			COALESCE(SUM(price), 0) as total_revenue,
			CASE 
				WHEN COUNT(*) = 0 THEN 0 
				ELSE ROUND((COUNT(CASE WHEN status = 'success' THEN 1 END) * 100.0 / COUNT(*)), 2)
			END as success_rate
		FROM transactions 
		WHERE created_at >= $1 AND created_at <= $2
	`

	row := repo.DB.QueryRow(query, startDate, endDate)

	var result AnalyticsResult
	err := row.Scan(
		&result.Date,
		&result.TotalTransactions,
		&result.TotalProfit,
		&result.TotalRevenue,
		&result.SuccessRate,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAnalyticsByStatus - Mendapatkan analytics berdasarkan status
func (repo *AnalyticsRepository) GetAnalyticsByStatus(startDate, endDate time.Time, status string) (*AnalyticsResult, error) {
	query := `
		SELECT 
			$3 as status,
			COUNT(*) as total_transactions,
			COALESCE(SUM(profit_amount), 0) as total_profit,
			COALESCE(SUM(price), 0) as total_revenue,
			100.0 as success_rate
		FROM transactions 
		WHERE created_at >= $1 AND created_at <= $2 AND status = $3
	`

	row := repo.DB.QueryRow(query, startDate, endDate, status)

	var result AnalyticsResult
	err := row.Scan(
		&result.Date,
		&result.TotalTransactions,
		&result.TotalProfit,
		&result.TotalRevenue,
		&result.SuccessRate,
	)

	if err == sql.ErrNoRows {
		return &AnalyticsResult{
			Date:              status,
			TotalTransactions: 0,
			TotalProfit:       0,
			TotalRevenue:      0,
			SuccessRate:       0,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}
