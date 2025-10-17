package di

import "sitex/internal/user"

type IUserRepository interface {
	GetEmployeeInfo(email string) (user.Employee, error)
	ChangePassword(email, password string) error
	EmailExists(email string) bool
	CreateUserWithDepartment(FirstName, LastName, Email, Department, Position, Password string, IsActive, IsAdmin, ShowTimeEvents bool) error
	IsAdmin(email string) bool
}
