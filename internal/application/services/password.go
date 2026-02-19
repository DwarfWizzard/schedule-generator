package services

type PasswordService interface {
	HashPassword(pwd string) (string, error)
	CompareWithHash(hashed, pwd string) bool
}
