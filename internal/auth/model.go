package auth

type LoginForm struct {
	Email    string
	Password string
}

type userUpdateForm struct {
	FirstName  string `form:"first_name"`
	LastName   string `form:"last_name"`
	Position   string `form:"position"`
	Department string `form:"department"`
	IsActive   string `form:"is_active"` // будет "true" если отмечен, либо пустая строка
	IsAdmin    string `form:"is_admin"`  // будет "true" если отмечен, либо пустая строка
	Email      string `form:"email"`
}
