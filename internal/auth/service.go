package auth

import (
	"errors"
	"sitex/pkg/di"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository di.IUserRepository
}

func NewAuthService(userRepository di.IUserRepository) *AuthService {
	return &AuthService{userRepository}
}

func (service *AuthService) Login(
	loginUserForm LoginForm,
) error {
	existedUser, err := service.UserRepository.GetEmployeeInfo(loginUserForm.Email)
	if err != nil {
		return errors.New(ErrWrongCredentials)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(existedUser.PasswordHash),
		[]byte(loginUserForm.Password),
	)
	if err != nil {
		return errors.New(ErrWrongCredentials)
	}

	return nil
}

func (service *AuthService) ChangePassword(
	changePasswordForm changePasswordForm,
) error {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(changePasswordForm.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	err = service.UserRepository.ChangePassword(
		changePasswordForm.Email,
		string(hashedPassword),
	)
	if err != nil {
		return err
	}
	return nil
}

func (service *AuthService) Register(
	form userCreateForm,
) error {
	existedUser := service.UserRepository.EmailExists(form.Email)
	if existedUser {
		return errors.New(ErrUserExists)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(form.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	// Конвертируем строки в boolean
	isActive := form.IsActive == "true"
	isAdmin := form.IsAdmin == "true"

	err = service.UserRepository.CreateUserWithDepartment(
		form.FirstName,
		form.LastName,
		form.Email,
		form.Department,
		form.Position,
		string(hashedPassword),
		isAdmin,
		isActive,
	)
	if err != nil {
		return err
	}
	return nil
}
