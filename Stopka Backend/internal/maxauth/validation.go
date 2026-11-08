
package maxauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

func Validate(initData, botToken string) (*User, error) {
	if initData == "" || botToken == "" {
		return nil, errors.New("отсутствуют данные авторизации")
	}

	if len(initData) > 16384 {
		return nil, errors.New("слишком длинные данные авторизации")
	}

	params := make(map[string]string)

	for _, part := range strings.Split(initData, "&") {
		pair := strings.SplitN(part, "=", 2)
		if len(pair) != 2 || pair[0] == "" {
			return nil, errors.New("неверный формат данных")
		}

		key, err := url.QueryUnescape(pair[0])
		if err != nil {
			return nil, errors.New("неверная кодировка ключа")
		}

		value, err := url.QueryUnescape(pair[1])
		if err != nil {
			return nil, errors.New("неверная кодировка значения")
		}

		if _, exists := params[key]; exists {
			return nil, errors.New("повторяющийся параметр")
		}

		params[key] = value
	}

	providedHash, ok := params["hash"]
	if !ok {
		return nil, errors.New("отсутствует подпись")
	}

	signature, err := hex.DecodeString(providedHash)
	if err != nil || len(signature) != sha256.Size {
		return nil, errors.New("неверный формат подписи")
	}

	delete(params, "hash")

	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, key+"="+params[key])
	}

	checkString := strings.Join(lines, "\n")

	first := hmac.New(sha256.New, []byte("WebAppData"))
	first.Write([]byte(botToken))
	secret := first.Sum(nil)

	second := hmac.New(sha256.New, secret)
	second.Write([]byte(checkString))
	expected := second.Sum(nil)

	if !hmac.Equal(signature, expected) {
		return nil, errors.New("подпись MAX не совпадает")
	}

	authDate, err := strconv.ParseInt(params["auth_date"], 10, 64)
	if err != nil || authDate <= 0 {
		return nil, errors.New("неверная дата авторизации")
	}

	issuedAt := time.Unix(authDate, 0)
	now := time.Now()

	if issuedAt.After(now.Add(time.Minute)) ||
		now.Sub(issuedAt) > 24*time.Hour {
		return nil, errors.New("данные авторизации устарели")
	}

	var user User
	if err := json.Unmarshal([]byte(params["user"]), &user); err != nil {
		return nil, errors.New("неверные данные пользователя")
	}

	if user.ID <= 0 {
		return nil, errors.New("неверный ID пользователя MAX")
	}

	return &user, nil
}
