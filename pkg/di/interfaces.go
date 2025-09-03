package di

import "sitex/internal/user"

type IUserRepository interface {
	GetEmployeeInfo(email string) (user.Employee, error)
	ChangePassword(email, password string) error
	EmailExists(email string) bool
	CreateUser(firstName, lastName, email, department, position, password string, isActive, isAdmin bool) error
}
