package di

import "sitex/internal/user"

type IUserRepository interface {
	GetEmployeeInfo(email string) (user.Employee, error)
	ChangePassword(email, password string) error
}
