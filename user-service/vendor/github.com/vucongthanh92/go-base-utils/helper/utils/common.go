package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"runtime/debug"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	mrand "math/rand"

	"github.com/spf13/viper"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type clientIPContextKey struct{}

// ----------------------------------------------------------------------
// FUNCTION: handling errors in goroutines
// ----------------------------------------------------------------------

// SafeGo runs the provided function in a new goroutine and recovers from any panic, logging the error.
func SafeGo(f func()) {
	go func() {
		defer HandlePanic()
		f()
	}()
}

// HandlePanic recovers from a panic and logs the error and stack trace.
func HandlePanic() {
	if r := recover(); r != nil {
		logger.Error("Recovered from panic: ", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
	}
}

// ----------------------------------------------------------------------
// FUNCTION: Processing strings and arrays using iteration and arbitrary types.
// ----------------------------------------------------------------------

// Reverse reverses a UTF-8 encoded string. It returns an error if the input is not valid UTF-8.
func Reverse(s string) (string, error) {
	if !utf8.ValidString(s) {
		return s, errors.New("input is not valid UTF-8")
	}
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r), nil
}

func IterateSlice[T any](params []T, f func(i int, item T)) {
	if len(params) == 0 {
		return
	}
	for i, item := range params {
		f(i, item)
	}
}

func IterateMap[T comparable, K any](params map[T]K, f func(i T, item K)) {
	if params == nil {
		return
	}
	for i, item := range params {
		f(i, item)
	}
}

func Contains[T comparable](values []T, target T) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func LowerInitial(fields []string) (results []string) {
	for _, str := range fields {
		result := ""
		for j, val := range str {
			result = string(unicode.ToLower(val)) + str[j+1:]
			break
		}
		results = append(results, result)
	}
	return results
}

func RemoveDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func SliceToMap[T any, K comparable](arr []T, f func(item T) (K, T)) map[K]T {
	result := make(map[K]T)
	for _, item := range arr {
		key, item := f(item)
		result[key] = item
	}
	return result
}

// ConvertStrToStruct converts a JSON string to a struct of type T.
// It returns an error if the conversion fails.
func ConvertStrToStruct[T any](param string) (T, error) {
	var resp T
	err := json.Unmarshal([]byte(param), &resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// ConvertStructToStr converts a struct to its JSON string representation.
func ConvertStructToStr[T any](param T) (string, error) {
	var resp string
	buff, err := json.Marshal(param)
	if err != nil {
		return resp, err
	}
	resp = string(buff)
	return resp, nil
}

// ----------------------------------------------------------------------
// FUNCTION: Handling encryption, hashing,
// and the generation of random strings such as keys and passwords.
// ----------------------------------------------------------------------

// HashPwdBySha256 hashes the password using
func HashPwdBySha256(email, password string) string {
	secret := viper.GetString("authenticate.passwordHashSecret")
	hashMethod := sha256.New()
	hashMethod.Write([]byte(secret + email + password))
	hash := hashMethod.Sum(nil)
	result := strings.ToUpper(hex.EncodeToString(hash))
	return result
}

// GetHeaderFromContext retrieves the header value from the context for the specified key.
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

// ParseUserID converts the sub claim to uint64 safely.
func ParseUserID(sub any) uint64 {
	switch v := sub.(type) {
	case string:
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	case float64:
		return uint64(v)
	case json.Number:
		id, _ := v.Int64()
		return uint64(id)
	default:
		return 0
	}
}
