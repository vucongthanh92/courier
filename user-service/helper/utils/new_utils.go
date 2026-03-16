package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"runtime/debug"
	"strings"
	"time"

	mrand "math/rand"

	"github.com/gin-gonic/gin"
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

func GetUserAgent(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if ginCtx, ok := ctx.(*gin.Context); ok {
		if ginCtx.Request != nil {
			return ginCtx.Request.UserAgent()
		}
		return ""
	}

	return GetHeaderFromKey(ctx, "headers", "User-Agent")
}

func GetClientIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.ClientIP()
	}

	if ip := getIPFromHeader(ctx, "headers"); ip != "" {
		return ip
	}

	return ""
}

// RandString generates a random string of the specified length using crypto/rand for better randomness.
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	mrand.Seed(time.Now().UnixNano())
}

func RandString(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, n)
	max := big.NewInt(int64(len(letters)))

	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err == nil {
			buf[i] = letters[v.Int64()]
			continue
		}
		// Fallback: non-crypto random
		buf[i] = letters[mrand.Intn(len(letters))]
	}
	return string(buf)
}
