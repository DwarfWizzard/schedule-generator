package services

type PasswordService interface {
	HashPassword(pwd string) string
	CompareWithHash(hashed, pwd string) bool
}
