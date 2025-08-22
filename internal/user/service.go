package user

type UserServiceDeps struct {
	UserRepository UserRepository
}

type UserService struct {
	userRepository UserRepository
}

func NewUserService(deps *UserServiceDeps) *UserService {
	return &UserService{userRepository: deps.UserRepository}
}

func (service *UserService) Days(email string, dateTo, dateFrom string) error {
	// days := service.userRepository.GetStatusHistory(email)
	return nil
}
