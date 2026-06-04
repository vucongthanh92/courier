package utils

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	httpreq "github.com/vucongthanh92/go-base-utils/http/request"
)

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

func GetHashMD5(input string) string {
	sum := md5.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}

func GetToken(id string) string {
	secret := os.Getenv("PASSWORD_SECRET")
	hashMethod := sha256.New()
	hashMethod.Write([]byte(secret + id))
	hash := hashMethod.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(hash))
}

func LowerInitial(fields []string) (results []string) {
	for _, str := range fields {
		for j, val := range str {
			results = append(results, string(unicode.ToLower(val))+str[j+1:])
			break
		}
	}
	return results
}

func Contains[T comparable](s []T, str T) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
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

func FillTemplate(templateJSON string, data map[string]string) string {
	filledJSON := templateJSON
	for key, value := range data {
		placeholder := "${" + key + "}"
		filledJSON = strings.ReplaceAll(filledJSON, placeholder, value)
	}
	return filledJSON
}

func SliceToMap[T any, K comparable](arr []T, f func(item T) (K, T)) map[K]T {
	result := make(map[K]T)
	for _, item := range arr {
		key, mapped := f(item)
		result[key] = mapped
	}
	return result
}

func IterateSlice[T any](params []T, f func(i int, item T)) {
	for i, item := range params {
		f(i, item)
	}
}

func IterateMap[T comparable, K any](params map[T]K, f func(i T, item K)) {
	for i, item := range params {
		f(i, item)
	}
}

func ConvertStrToStruct[T any](param string) (T, error) {
	var resp T
	err := json.Unmarshal([]byte(param), &resp)
	return resp, err
}

func ConvertStructToStr[T any](param T) (string, error) {
	buff, err := json.Marshal(param)
	if err != nil {
		return "", err
	}
	return string(buff), nil
}

func CheckDuplicateTypeByField[T any](options []T, field string) (err error, isDuplicate bool) {
	allKeys := make(map[interface{}]bool)
	for _, item := range options {
		val := reflect.ValueOf(item)
		fieldValue := val.FieldByName(field)
		if !fieldValue.IsValid() {
			return fmt.Errorf("field %s does not exist in struct", field), false
		}
		key := fieldValue.Interface()
		if _, exists := allKeys[key]; exists {
			return nil, true
		}
		allKeys[key] = true
	}
	return nil, false
}

func GetHeaderFromKey(ctx context.Context, key, field string) (resp string) {
	headers := httpreq.GetHeaderFromContext(ctx, key)
	if values := headers[field]; len(values) > 0 {
		return values[0]
	}
	return ""
}
