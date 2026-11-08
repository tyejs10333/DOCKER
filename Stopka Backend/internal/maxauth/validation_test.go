
package maxauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func makeSignedData(t *testing.T, token string, userID int64, authDate int64) string {
	t.Helper()

	userJSON, err := json.Marshal(map[string]any{
		"id":         userID,
		"first_name": "Test",
	})
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(authDate, 10))
	values.Set("user", string(userJSON))

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, key+"="+values.Get(key))
	}

	first := hmac.New(sha256.New, []byte("WebAppData"))
	first.Write([]byte(token))

	second := hmac.New(sha256.New, first.Sum(nil))
	second.Write([]byte(strings.Join(lines, "\n")))

	values.Set("hash", hex.EncodeToString(second.Sum(nil)))
	return values.Encode()
}

func TestValidateValidData(t *testing.T) {
	data := makeSignedData(t, "test-bot-token", 12345, time.Now().Unix())

	user, err := Validate(data, "test-bot-token")
	if err != nil {
		t.Fatalf("Подлинные данные отклонены: %v", err)
	}
	if user.ID != 12345 {
		t.Fatalf("Неверный пользователь: %d", user.ID)
	}
}

func TestValidateWrongToken(t *testing.T) {
	data := makeSignedData(t, "correct-token", 12345, time.Now().Unix())

	if _, err := Validate(data, "wrong-token"); err == nil {
		t.Fatal("Данные с неправильным токеном были приняты")
	}
}

func TestValidateModifiedUser(t *testing.T) {
	data := makeSignedData(t, "test-token", 12345, time.Now().Unix())
	values, err := url.ParseQuery(data)
	if err != nil {
		t.Fatal(err)
	}

	values.Set("user", `{"id":99999,"first_name":"Attacker"}`)

	if _, err := Validate(values.Encode(), "test-token"); err == nil {
		t.Fatal("Подмена пользователя не обнаружена")
	}
}

func TestValidateExpiredData(t *testing.T) {
	data := makeSignedData(
		t,
		"test-token",
		12345,
		time.Now().Add(-25*time.Hour).Unix(),
	)

	if _, err := Validate(data, "test-token"); err == nil {
		t.Fatal("Устаревшие данные были приняты")
	}
}

func TestValidateDuplicateParameter(t *testing.T) {
	data := makeSignedData(t, "test-token", 12345, time.Now().Unix())
	data += "&auth_date=123"

	if _, err := Validate(data, "test-token"); err == nil {
		t.Fatal("Повторяющийся параметр был принят")
	}
}
