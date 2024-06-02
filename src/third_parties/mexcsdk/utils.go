package mexcsdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func JsonToParamStr(jsonParams string) string {
	var paramsarr []string
	var arritem string
	m := make(map[string]string)
	err := json.Unmarshal([]byte(jsonParams), &m)
	if err != nil {
		return ""
	}
	i := 0
	for key, value := range m {

		arritem = fmt.Sprintf("%s=%s", key, value)
		paramsarr = append(paramsarr, arritem)
		i++
		if i > len(m) {
			break
		}
	}
	paramsstr := strings.Join(paramsarr, "&")
	return paramsstr
}

func paramsEncode(paramStr string) string {
	return url.QueryEscape(paramStr)
}

func computeHmac256(message string, secKey string) string {
	key := []byte(secKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func sign(url, jsonParams, secretKey string) string {
	timestamp := int64(float64(time.Now().UnixMilli()))
	path := ""
	if jsonParams == "" {
		message := fmt.Sprintf("timestamp=%d", timestamp)
		sign := computeHmac256(message, secretKey)
		path = fmt.Sprintf("%s?timestamp=%d&signature=%s", url, timestamp, sign)
	} else {
		strParams := JsonToParamStr(jsonParams)
		message := fmt.Sprintf("%s&timestamp=%d", strParams, timestamp)
		sign := computeHmac256(message, secretKey)
		path = fmt.Sprintf("%s?%s&timestamp=%d&signature=%s", url, strParams, timestamp, sign)
	}

	return path
}

func signQuery(url, queryParams, secretKey string) string {
	timestamp := int64(float64(time.Now().UnixMilli()))
	message := fmt.Sprintf("%s&timestamp=%d", queryParams, timestamp)
	sign := computeHmac256(message, secretKey)
	path := fmt.Sprintf("%s?%s&timestamp=%d&signature=%s", url, queryParams, timestamp, sign)

	return path
}
