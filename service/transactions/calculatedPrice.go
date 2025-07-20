package transactions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type CalculatedPriceResponse struct {
	Price  int
	Profit int
}

type UserRole string

const (
	RolePlatinum UserRole = "PLATINUM"
	RoleAdmin    UserRole = "ADMIN"
	RoleMember   UserRole = "MEMBER"
)

type ServicePricing struct {
	PlatinumPrice  int
	ResellerPrice  int
	MemberPrice    int
	PlatinumProfit int
	ResellerProfit int
	MemberProfit   int
}

func (repo TransactionsRepository) CalculatedPriceAndProfit(c context.Context, username string, productCode string) (*CalculatedPriceResponse, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	if productCode == "" {
		return nil, errors.New("product code cannot be empty")
	}

	role, err := repo.getUserRole(c, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	servicePricing, err := repo.getServicePricing(c, productCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get service pricing: %w", err)
	}

	price, profit := repo.calculatePriceAndProfitByRole(role, servicePricing)

	return &CalculatedPriceResponse{
		Price:  price,
		Profit: profit,
	}, nil
}

func (repo TransactionsRepository) getUserRole(c context.Context, username string) (UserRole, error) {
	var role string
	query := `SELECT role FROM users WHERE username = $1`

	err := repo.DB.QueryRowContext(c, query, username).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("user not found")
		}
		return "", fmt.Errorf("database query failed: %w", err)
	}

	return UserRole(role), nil
}

func (repo TransactionsRepository) getServicePricing(c context.Context, productCode string) (*ServicePricing, error) {
	var pricing ServicePricing

	query := `
	SELECT 
		platinum_price,
		reseller_price,
		member_price,
		platinum_profit,
		reseller_profit,
		member_profit
	FROM 
		services
	WHERE 
		provider_id = $1
	`

	err := repo.DB.QueryRowContext(c, query, productCode).Scan(
		&pricing.PlatinumPrice,
		&pricing.ResellerPrice,
		&pricing.MemberPrice,
		&pricing.PlatinumProfit,
		&pricing.ResellerProfit,
		&pricing.MemberProfit,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("service not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	return &pricing, nil
}

func (repo TransactionsRepository) calculatePriceAndProfitByRole(role UserRole, pricing *ServicePricing) (int, int) {
	switch role {
	case RolePlatinum:
		return pricing.PlatinumPrice, pricing.PlatinumProfit
	case RoleAdmin:
		return pricing.ResellerPrice, pricing.ResellerProfit
	case RoleMember:
		return pricing.MemberPrice, pricing.MemberProfit
	default:
		return pricing.MemberPrice, pricing.MemberProfit
	}
}
