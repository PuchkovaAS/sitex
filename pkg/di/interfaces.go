package di

import "sitex/internal/user"

type IUserRepository interface {
	FindByEmail(email string) (*user.Employee, error)
}
