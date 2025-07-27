// analytics_service.go
package analytics

import (
	"errors"
	"time"
)

type AnalyticsService struct {
	repo *AnalyticsRepository
}

func NewAnalyticsService(repo *AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		repo: repo,
	}
}

func (s *AnalyticsService) GetAnalyticsByDateRange(startDateStr, endDateStr string) ([]AnalyticsResult, error) {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	return s.repo.GetAnalyticsByDateRange(startDate, endDate)
}

func (s *AnalyticsService) GetAnalyticsByDate(dateStr string) (*AnalyticsResult, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	return s.repo.GetAnalyticsByDate(date)
}

func (s *AnalyticsService) GetTodayAnalytics() (*AnalyticsResult, error) {
	return s.repo.GetTodayAnalytics()
}

func (s *AnalyticsService) GetMonthlyAnalytics(year int, month int) ([]AnalyticsResult, error) {
	if month < 1 || month > 12 {
		return nil, errors.New("invalid month, must be between 1-12")
	}

	if year < 2020 || year > time.Now().Year()+1 {
		return nil, errors.New("invalid year")
	}

	return s.repo.GetMonthlyAnalytics(year, time.Month(month))
}

func (s *AnalyticsService) GetAnalyticsSummary(startDateStr, endDateStr string) (*AnalyticsResult, error) {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	return s.repo.GetAnalyticsSummary(startDate, endDate)
}

func (s *AnalyticsService) GetAnalyticsByStatus(startDateStr, endDateStr, status string) (*AnalyticsResult, error) {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	if status == "" {
		return nil, errors.New("status cannot be empty")
	}

	return s.repo.GetAnalyticsByStatus(startDate, endDate, status)
}
