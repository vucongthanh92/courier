package httpcommon

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vucongthanh92/go-base-project/helper/constants"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

func ParseValue(s string) any {
	// parse bool
	if val, err := strconv.ParseBool(s); err == nil {
		return val
	}
	// parse float
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	// parse int
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val
	}
	return s
}

func ParseQueryParams(params map[string][]string) map[string]any {
	parsedParams := make(map[string]any)
	for k, v := range params {
		if k == "pageSize" || k == "pageIndex" || k == "fromDate" || k == "toDate" {
			continue
		}
		if len(v) == 1 {
			if v[0] != "" {
				parsedParams[k] = ParseValue(v[0])
			}
		} else {
			parsedParams[k] = v
		}
	}
	return parsedParams
}

func GetParamObjectID(c *gin.Context) (int64, error) {

	//get jwt token
	token := GetAuthToken(c)
	if token == "" {
		return 0, errors.New(constants.TOKEN_MISSING)
	}

	// parse jwt without verify
	data, err := parseJwt(token)
	if err != nil {
		return 0, err
	}

	// get user id
	return int64(data["_id"].(float64)), nil
}

func parseJwt(jwtToken string) (jwt.MapClaims, error) {
	tokenString := strings.Replace(jwtToken, "Bearer ", "", 1)
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return token, nil
	})
	return token.Claims.(jwt.MapClaims), nil
}

func GetAdminID(c *gin.Context) (int64, error) {
	//get jwt token
	token := GetAuthToken(c)
	if token == "" {
		return 0, errors.New(constants.TOKEN_MISSING)
	}
	// parse jwt without verify
	data, err := parseJwt(token)
	if err != nil {
		return 0, err
	}

	adminID, isValid := data["_adminID"].(float64)
	if !isValid {
		return 0, errors.New("cannot parse params")
	}

	resp := int64(adminID)
	if reflect.ValueOf(resp).IsZero() {
		return 0, errors.New(constants.TOKEN_MISSING)
	}

	// get user id
	return resp, nil
}

func GetAuthToken(c *gin.Context) (authToken string) {
	authToken = c.Request.Header.Get("Authorization")
	if authToken == "" {
		var err error
		authToken, err = c.Cookie("access_token")
		if err != nil {
			logger.Error("GetAuthToken cookie err: ", zap.Error(err))
			return ""
		}
	}

	return authToken
}
