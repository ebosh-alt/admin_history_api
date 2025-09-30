package misc

import "golang.org/x/crypto/bcrypt"

// HashPassword генерирует bcrypt-хэш из переданного пароля.
func HashPassword(rawPassword string, pepper string) (string, error) {

	passwordWithPepper := rawPassword + pepper
	const cost = 12

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), cost)
	if err != nil {
		return "", err
	}

	// 4. Возвращаем строку: bcrypt-хэш содержит соль и cost внутри себя
	return string(hashBytes), nil
}

// ComparePassword сравнивает переданный пароль с хэшем из базы данных.
func ComparePassword(passwordHash string, password string, pepper string) error {
	passwordWithPepper := password + pepper

	// Сравниваем хэш и пароль
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(passwordWithPepper))
}
