package auth

// import (
// 	"errors"
// 	"sitex/pkg/di"
//
// 	"golang.org/x/crypto/bcrypt"
// )
//
// type AuthService struct {
// 	UserRepository di.IUserRepository
// }
//
// func NewAuthService(userRepository di.IUserRepository) *AuthService {
// 	return &AuthService{userRepository}
// }

// func (service *AuthService) Login(
// 	loginUserForm userLoginForm,
// ) error {
// 	existedUser, _ := service.UserRepository.FindByEmail(loginUserForm.Email)
//
// 	if existedUser == nil {
// 		return errors.New(ErrWrongCredentials)
// 	}
//
// 	err := bcrypt.CompareHashAndPassword(
// 		[]byte(existedUser.Password),
// 		[]byte(loginUserForm.Password),
// 	)
// 	if err != nil {
// 		return errors.New(ErrWrongCredentials)
// 	}
//
// 	return nil
// }
