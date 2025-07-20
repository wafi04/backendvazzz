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

	// Find user
	user, err := service.userRepo.FindUserByUsername(data.Username)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid username or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := config.GenerateJWT(strconv.Itoa(user.Id), user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	// Update user token in database
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
func (service *UserService) validateRegisterRequest(req *model.RegisterRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if req.Whatsapp == "" {
		return errors.New("whatsapp is required")
	}
	return nil
}
