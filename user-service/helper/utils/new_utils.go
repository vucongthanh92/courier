package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"runtime/debug"
	"strings"

	"github.com/spf13/viper"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func SafeGo(f func()) {
	go func() {
		defer HandlePanic()
		f()
	}()
}

func HandlePanic() {
	if r := recover(); r != nil {
		logger.Error("Recovered from panic: ", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
	}
}

// HashPwdByBcrypt hashes the password using bcrypt algorithm.
func HashPwdByBcrypt(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}

// CheckPwdByBcrypt compares the hashed password with its possible plaintext equivalent.
func CheckPwdByBcrypt(hash, pwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
}

// HashPwdBySha256 hashes the password using
func HashPwdBySha256(email, password string) string {
	secret := viper.GetString("authenticate.passwordHashSecret")
	hashMethod := sha256.New()
	hashMethod.Write([]byte(secret + email + password))
	hash := hashMethod.Sum(nil)
	result := strings.ToUpper(hex.EncodeToString(hash))
	return result
}
