package users

import (
	"errors"
	"strconv"

	"github.com/wafi04/backendvazzz/pkg/config"
	"github.com/wafi04/backendvazzz/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *UserRepository
}

func NewUserService(userRepo *UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (service *UserService) Create(data model.RegisterRequest) (*model.UserData, error) {

	existingUser, err := service.userRepo.FindUserByUsername(data.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	return service.userRepo.Register(&data)
}

func (service *UserService) Login(data model.LoginRequest) (*model.UserData, string, error) {
	// Validate input
	if data.Username == "" || data.Password == "" {
		return nil, "", errors.New("username and password are required")
	}

	user, err := service.userRepo.FindUserByUsername(data.Username)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	token, err := config.GenerateJWT(strconv.Itoa(user.Id), user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	err = service.userRepo.UpdateUserToken(strconv.Itoa(user.Id), token)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (service *UserService) GetProfile(username string) (*model.UserData, error) {
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (service *UserService) UpdateBalance(userID string, newBalance int) error {
	return service.userRepo.UpdateUserBalance(userID, newBalance)
}

func (service *UserService) UpdateLastPayment(userID string) error {
	return service.userRepo.UpdateLastPayment(userID)
}

func (service *UserService) Logout(userID string) error {
	return service.userRepo.UpdateUserToken(userID, "")
}
