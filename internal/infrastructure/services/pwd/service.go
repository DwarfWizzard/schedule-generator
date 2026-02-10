package pwd

import "golang.org/x/crypto/bcrypt"

type service struct {
	passwordSalt string
}

func NewPasswordService(salt string) *service {
	return &service{
		passwordSalt: salt,
	}
}

// HashPassword
func (s *service) HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CompareWithHash
func (s *service) CompareWithHash(hashed, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pwd)) == nil
}
