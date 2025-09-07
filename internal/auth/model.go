package auth

type LoginForm struct {
	Email    string
	Password string
}

type changePasswordForm struct {
	Email           string `form:"email"`
	NewPassword     string `form:"new_password"`
	ConfirmPassword string `form:"confirm_password"`
}

type userUpdateForm struct {
	FirstName     string `form:"first_name"`
	LastName      string `form:"last_name"`
	Position      string `form:"position"`
	Department    string `form:"department_id"`
	IsActive      string `form:"is_active"` // будет "true" если отмечен, либо пустая строка
	IsAdmin       string `form:"is_admin"`  // будет "true" если отмечен, либо пустая строка
	Email         string `form:"email"`
	NewDepartment string `form:"new_department"`
}

type userCreateForm struct {
	FirstName       string `form:"first_name"`
	LastName        string `form:"last_name"`
	Position        string `form:"position"`
	Department      string `form:"department_id"`
	NewDepartment   string `form:"new_department"`
	IsActive        string `form:"is_active"` // будет "true" если отмечен, либо пустая строка
	IsAdmin         string `form:"is_admin"`  // будет "true" если отмечен, либо пустая строка
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}
