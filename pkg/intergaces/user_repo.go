package intergaces

type IUserRepository interface {
	IsAdmin(email string) bool
}
