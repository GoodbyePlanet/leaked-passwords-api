package service

type CheckPassword struct {
}

func NewCheckPassword() *CheckPassword {
	return &CheckPassword{}
}

func (checkPassword *CheckPassword) CheckPassword(password string) bool {
	return password == "123"
}
