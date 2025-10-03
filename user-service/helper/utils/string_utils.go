package utils

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	cacheV9 "github.com/go-redis/cache/v9"
	"github.com/spf13/viper"
	"github.com/vucongthanh92/go-base-utils/cache"
	utilsSvc "github.com/vucongthanh92/go-base-utils/http/request"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
	"gopkg.in/guregu/null.v3"
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

func HashPassword(email, password string) string {
	secret := viper.GetString("authenticate.passwordHashSecret")
	hashMethod := sha256.New()
	hashMethod.Write([]byte(secret + email + password))
	hash := hashMethod.Sum(nil)
	result := strings.ToUpper(hex.EncodeToString(hash))
	return result
}

func GetHashMD5(phone string) string {
	hashMethod := md5.New()
	hashMethod.Write([]byte(phone))
	hash := hashMethod.Sum(nil)
	return hex.EncodeToString(hash)
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
		result := ""
		for j, val := range str {
			result = string(unicode.ToLower(val)) + str[j+1:]
			break
		}
		results = append(results, result)
	}
	return results
}

// https://play.golang.org/p/Qg_uv_inCek
// contains checks if a string is present in a slice
func Contains[T comparable](s []T, str T) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func ParserSeoulTime(val time.Time) time.Time {
	return val.UTC().Add(9 * time.Hour)
}

func FormatSeoulTime(val time.Time) time.Time {
	locSeoul := time.FixedZone("", 9*60*60)

	resp := time.Date(val.Year(), val.Month(), val.Day(),
		val.Hour(), val.Minute(), val.Second(), val.Nanosecond(), locSeoul,
	)

	return resp
}

func ConvertSeoulTime(val time.Time) time.Time {
	// Load the location for Seoul
	seoulLocation, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return time.Time{}
	}

	// Convert the appointment time to Seoul time
	seoulTime := val.In(seoulLocation)

	return RemoveNanoMilisecondTime(seoulTime)
}

func RemoveNanoMilisecondTime(val time.Time) time.Time {
	result := val.Format(time.RFC3339)
	value, _ := time.Parse(time.RFC3339, result)

	return value
}

func getSecondTimeCustom(val time.Time) int {
	hour := val.Hour()
	minute := val.Minute()
	second := val.Second()
	return hour*3600 + minute*60 + second
}

func AfterTimeCustom(timeFirst time.Time, timeSecond time.Time) bool {
	timeFirstVal := getSecondTimeCustom(timeFirst)
	timeSecondVal := getSecondTimeCustom(timeSecond)

	return timeFirstVal > timeSecondVal
}

func BeforeTimeCustom(timeFirst time.Time, timeSecond time.Time) bool {
	timeSecondVal := getSecondTimeCustom(timeSecond)
	timeFirstVal := getSecondTimeCustom(timeFirst)

	return timeFirstVal < timeSecondVal
}

func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}

func RoundDownPrice(num float64, precision int64) float64 {
	return float64(int64(num) - int64(num)%(10*(precision)))
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
		key, item := f(item)
		result[key] = item
	}
	return result
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

func ConvertStrToStruct[T any](param string) (T, error) {
	var resp T
	err := json.Unmarshal([]byte(param), &resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func ConvertStructToStr[T any](param T) (string, error) {
	var resp string
	buff, err := json.Marshal(param)
	if err != nil {
		return resp, err
	}
	resp = string(buff)
	return resp, nil
}

// check order_history_status

func CheckDuplicateTypeByField[T interface{}](options []T, field string) (err error, isDuplicate bool) {
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

func GetQueryConditionClauseInStruct(param any, queryArgs map[string]any) []string {
	var (
		resp     = make([]string, 0)
		elTypes  = reflect.TypeOf(param)
		elValues = reflect.ValueOf(param)
	)

	if elTypes.NumField() == 0 {
		return resp
	}

	for i := 0; i < elTypes.NumField(); i++ {
		var (
			elValue        = elValues.Field(i)
			elField        = elTypes.Field(i)
			queryClauseTag = elField.Tag.Get("query_clause")
			queryFieldTag  = elField.Tag.Get("query_field")
			valueArgs      any
		)

		if queryClauseTag == "" {
			continue
		}

		switch {
		case elField.Type.Kind() == reflect.String:
			{
				if elValue.String() == "" {
					continue
				}
				valueArgs = elValue.String()
			}
		case elField.Type.Kind() == reflect.Slice || elField.Type.Kind() == reflect.Array:
			{
				if elValue.Len() == 0 {
					continue
				}

				tempArr := make([]string, 0)
				for i := 0; i < elValue.Len(); i++ {
					if elValue.Index(i).Interface() == nil {
						continue
					}
					tempArr = append(tempArr, fmt.Sprintf("%v", elValue.Index(i).Interface()))
				}

				if len(tempArr) == 0 {
					continue
				}

				resp = append(resp, strings.ReplaceAll(queryClauseTag, "query_field", strings.Join(tempArr, ",")))
			}
		case elField.Type == reflect.TypeOf(null.String{}):
			{
				if !elValue.Interface().(null.String).Valid {
					continue
				}
				valueArgs = elValue.Interface().(null.String).String
			}
		case elField.Type == reflect.TypeOf(null.Int{}):
			{
				if !elValue.Interface().(null.Int).Valid {
					continue
				}
				valueArgs = elValue.Interface().(null.Int).Int64
			}
		case elField.Type == reflect.TypeOf(null.Float{}):
			{
				if !elValue.Interface().(null.Float).Valid {
					continue
				}
				valueArgs = elValue.Interface().(null.Float).Float64
			}
		case elField.Type == reflect.TypeOf(null.Time{}):
			{
				if !elValue.Interface().(null.Time).Valid {
					continue
				}
				valueArgs = elValue.Interface().(null.Time).Time
			}
		case elField.Type == reflect.TypeOf(null.Bool{}):
			{
				if !elValue.Interface().(null.Bool).Valid {
					continue
				}
				valueArgs = elValue.Interface().(null.Bool).Bool
			}
		default:
			valueArgs = elValue.Interface()
		}

		if valueArgs == nil {
			continue
		}

		resp = append(resp, strings.ReplaceAll(queryClauseTag, "query_field", ":"+queryFieldTag))
		queryArgs[queryFieldTag] = valueArgs
	}

	return resp
}

func IterateSlicePtr[T any](params []T, f func(i int, item *T)) {
	if len(params) == 0 {
		return
	}
	for i := range params {
		f(i, &params[i])
	}
}

func RandomIntInRange[T int | int16 | int32 | int64](start, end T) (res T, err error) {
	if end <= start {
		return res, errors.New("end value must be greater than start value")
	}

	// Calculate the range size.
	rangeSize := end - start + 1

	// Generate a random number in the range [0, rangeSize).
	randomBigInt, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
	if err != nil {
		return 0, err
	}

	// Convert the randomBigInt to an integer and shift it by the start value.
	randomInt := T(randomBigInt.Int64()) + start

	return randomInt, nil
}

func GetQueryCache[T any](cache cache.CacheInterface[string], ctx context.Context, key string, data T) (err error) {
	derivedCtx := context.WithValue(context.Background(), logger.TraceKey, ctx.Value(logger.TraceKey))
	val, err := cache.Get(derivedCtx, key)
	if err != nil {
		if err != cacheV9.ErrCacheMiss {
			logger.WarnCtx(derivedCtx, "GetQueryCache Error", zap.Error(err))
		}
		return err
	}
	return json.Unmarshal([]byte(*val), data)
}

func SetQueryCache[T any](cache cache.CacheInterface[string], ctx context.Context, key string, duration time.Duration, data T) {
	derivedCtx := context.WithValue(context.Background(), logger.TraceKey, ctx.Value(logger.TraceKey))

	valSave, _ := json.Marshal(data)
	errRedis := cache.Set(derivedCtx, key, string(valSave), duration)
	if errRedis != nil {
		logger.WarnCtx(derivedCtx, "SetQueryCache Error", zap.Error(errRedis))
		return
	}
}

func AlmostEqualFloat64(a, b float64) bool {
	const float64EqualityThreshold = 1e-6

	return math.Abs(a-b) <= float64EqualityThreshold
}

func GetHttpRequest[T any](req *http.Request) (response T, err error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("DefaultClient Error", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Warn("StatusCode Error", zap.Int("StatusCode", resp.StatusCode))
		return response, fmt.Errorf(resp.Status)
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ReadAll Error", zap.Error(err))
		return
	}
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		logger.Warn("Unmarshal Error", zap.Error(err))
		return response, nil
	}

	return response, nil
}

func RetryWithBackoff(attempts int, initialBackoff time.Duration, fn func() error) (lastAttemptErr error) {
	backoff := initialBackoff
	for i := 0; i < attempts; i++ {
		lastAttemptErr = fn()
		if lastAttemptErr == nil {
			return nil
		}
		logger.Warn("Attempt %d failed: %s\n" + lastAttemptErr.Error())

		if i < attempts-1 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return lastAttemptErr
}

func SetHeaderByKey(c *gin.Context, key string) context.Context {
	return utilsSvc.SetHeaderToContext(c, key)
}

func GetHeaderFromKey(ctx context.Context, key, field string) (resp string) {
	if key == "" || field == "" {
		return resp
	}

	headerMap := utilsSvc.GetHeaderFromContext(ctx, key)
	if val, existed := headerMap[field]; existed {
		resp = val[0]
	}

	return resp
}

func CompareEqualFold(src string, dst ...string) (resp bool) {
	if len(dst) == 0 {
		return resp
	}

	for _, val := range dst {
		if strings.EqualFold(src, val) {
			resp = true
			break
		}
	}

	return resp
}
