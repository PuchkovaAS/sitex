package auth

const (
	ErrUserCreate       = "Невозможно создать пользователя"
	ErrUserExists       = "Пользователь уже существует"
	ErrWrongCredentials = "Неверный email или пароль"
	ErrPasswordIsLess   = "Пароль должен содержать минимум 6 символов"
	ErrPasswordNotEq    = "Пароли не совпадают"
)
