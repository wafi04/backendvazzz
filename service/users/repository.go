package users

import (
	"database/sql"
	"errors"
	"time"

	"github.com/wafi04/backendvazzz/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(DB *sql.DB) *UserRepository {
	return &UserRepository{
		DB: DB,
	}
}

func (repo *UserRepository) FindUserByUsername(username string) (*model.UserData, error) {
	query := `
	SELECT
		id,
		name,
		username,
		whatsapp,
		balance,
		role,
		password,
		created_at,
		updated_at,
		last_Payment_at
	FROM 
		users
	WHERE 
		username = $1
	`

	row := repo.DB.QueryRow(query, username)

	var user model.UserData
	err := row.Scan(
		&user.Id,
		&user.Name,
		&user.Username,
		&user.Whatsapp,
		&user.Balance,
		&user.Role,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastPaymentAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) Register(req *model.RegisterRequest) (*model.UserData, error) {

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate UUID for user ID
	now := time.Now()

	query := `
	INSERT INTO users (
		name, 
		username, 
		password, 
		whatsapp, 
		balance, 
		role, 
		created_at, 
		updated_at, 
		last_payment_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING id, name, username, whatsapp, balance, role, created_at, updated_at, last_payment_at
	`

	var user model.UserData
	err = repo.DB.QueryRow(
		query,
		req.Name,
		req.Username,
		string(hashedPassword),
		req.Whatsapp,
		0,        // Default balance
		"MEMBER", // Default role
		now,      // createdAt
		now,      // updatedAt
		now,      // lastPaymentAt
	).Scan(
		&user.Id,
		&user.Name,
		&user.Username,
		&user.Whatsapp,
		&user.Balance,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastPaymentAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) Login(req *model.LoginRequest) (*model.UserData, error) {
	// Find user by username
	user, err := repo.FindUserByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	user.Password = ""

	return user, nil
}

func (repo *UserRepository) UpdateUserToken(userID, token string) error {
	query := `UPDATE users SET token = $1, updated_at = $2 WHERE id = $3`

	_, err := repo.DB.Exec(query, token, time.Now(), userID)
	return err
}

func (repo *UserRepository) FindUserByID(userID string) (*model.UserData, error) {
	query := `
	SELECT
		id,
		name,
		username,
		whatsapp,
		balance,
		role,
		otp,
		token,
        created_at,
		updated_at,
		last_payment_at
	FROM 
		users
	WHERE 
		id = $1
	`

	row := repo.DB.QueryRow(query, userID)

	var user model.UserData
	err := row.Scan(
		&user.Id,
		&user.Name,
		&user.Username,
		&user.Whatsapp,
		&user.Balance,
		&user.Role,
		&user.Otp,
		&user.Token,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastPaymentAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) UpdateUserBalance(userID string, newBalance int) error {
	query := `UPDATE users SET balance = $1, updated_at = $2 WHERE id = $3`

	_, err := repo.DB.Exec(query, newBalance, time.Now(), userID)
	return err
}

func (repo *UserRepository) UpdateLastPayment(userID string) error {
	query := `UPDATE users SET last_payment_at = $1, updated_at = $2 WHERE id = $3`

	_, err := repo.DB.Exec(query, time.Now(), time.Now(), userID)
	return err
}
