package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

const SignName = "sign"

func VerifySign(oSign string, params interface{}, secretKey string) error {
	if oSign == "" {
		return errors.New("sign not exit")
	}

	nSign, err := GetSign(params, secretKey)
	if err != nil || nSign != oSign {
		return errors.New("sign error")
	}
	return nil
}

func GetSign(params interface{}, secretKey string) (string, error) {
	m := make(map[string]interface{})
	if op, ok := params.(string); ok {
		m["default"] = op
	} else {
		data, err := json.Marshal(params)
		if err != nil {
			return "", errors.New("sign params not conform")
		}
		_ = json.Unmarshal(data, &m)
	}

	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	base := ""
	for _, k := range keys {
		base = base + fmt.Sprintf("%v%v", k, m[k])
	}
	base = base + secretKey
	h := md5.New()
	h.Write([]byte(base))
	return hex.EncodeToString(h.Sum(nil)), nil
}
