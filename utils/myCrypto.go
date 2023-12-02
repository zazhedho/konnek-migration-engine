package utils

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const encryptKey = "eyIwIjoia2xtbm8yIiwiMSI6IlBRUlNUOCIsIjIiOiJGR0hJSjYiLCIzIjoiYWJjZGUwIiwiNCI6InBxcnN0MyIsIjUiOiJBQkNERTUiLCI2IjoiZmdoaWoxIiwiNyI6IlVWV1hZWjkiLCI4IjoiS0xNTk83IiwiOSI6InV2d3h5ejQifQ"

// const encryptKey = "eyIwIjoia2xtbm8iLCIxIjoiUFFSU1QiLCIyIjoiRkdISUoiLCIzIjoiYWJjZGUiLCI0IjoicHFyc3QiLCI1IjoiQUJDREUiLCI2IjoiZmdoaWoiLCI3IjoiVVZXWFlaIiwiOCI6IktMTU5PIiwiOSI6InV2d3h5eiJ9"
// const encryptKey = "eyIwIjoiMiIsIjEiOiI4IiwiMiI6IjYiLCIzIjoiMCIsIjQiOiIzIiwiNSI6IjUiLCI2IjoiMSIsIjciOiI5IiwiOCI6IjciLCI5IjoiNCJ9"
type MyCrypto struct {
	encryptMap  map[int]string
	Expired     int64       `json:"expired"`
	Data        interface{} `json:"data"`
	EncryptedAt time.Time   `json:"encrypted_at"`
	EncryptedBy string      `json:"encrypted_by"`
}

func NewCrypto(key string) (*MyCrypto, error) {
	if key == "" {
		key = encryptKey
	}

	var em map[int]string
	seg := ""
	if l := len(key) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	if v, err := base64.URLEncoding.DecodeString(key + seg); err != nil {
		return nil, err
	} else if err = json.Unmarshal(v, &em); err != nil {
		return nil, err
	}

	return &MyCrypto{
		encryptMap: em,
	}, nil
}

func (mc *MyCrypto) randCode(length, tipe int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	switch tipe {
	case 1:
		chars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz")
		break
	case 2:
		chars = []rune("0123456789")
	}

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}

	return b.String()
}

func (mc *MyCrypto) Encrypt(data interface{}, ttl int64, encryptedBy string) (string, int64, error) {
	rand.Seed(time.Now().UnixNano())
	lSalt := mc.randCode(rand.Intn(9-3)+3, 2)

	time.Sleep(time.Nanosecond * 2)
	rand.Seed(time.Now().UnixNano())
	rSalt := mc.randCode(rand.Intn(9-3)+3, 2)

	if ttl > 0 {
		ttl = time.Now().Add(time.Duration(ttl) * time.Second).Unix()
	}

	b, err := json.Marshal(MyCrypto{
		Expired:     ttl,
		Data:        data,
		EncryptedAt: time.Now(),
		EncryptedBy: encryptedBy,
	})
	if err != nil {
		return "", ttl, err
	}

	hash := fmt.Sprintf("%s%s%s%d%d", lSalt, hex.EncodeToString(b), rSalt, len(rSalt), len(lSalt))
	eachPart := int(math.Floor(float64(len(hash)) / 10))
	kParts := make([]int, 9)
	vParts := make([]string, 10)
	for i := 0; i < 9; i++ {
		vParts[i] = hash[(i * eachPart) : (i*eachPart)+eachPart]
		kParts[i] = i
	}
	vParts[9] = hash[9*eachPart:]

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(kParts), func(i, j int) { kParts[i], kParts[j] = kParts[j], kParts[i] })

	k := ""
	v := ""
	for i := 0; i < len(kParts); i++ {
		secretKeys := strings.Split(mc.encryptMap[kParts[i]], "")

		time.Sleep(time.Nanosecond * 2)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(secretKeys), func(i, j int) { secretKeys[i], secretKeys[j] = secretKeys[j], secretKeys[i] })

		v += vParts[kParts[i]]
		k += secretKeys[0]
	}

	secretKeys := strings.Split(mc.encryptMap[len(kParts)], "")

	time.Sleep(time.Nanosecond * 2)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(secretKeys), func(i, j int) { secretKeys[i], secretKeys[j] = secretKeys[j], secretKeys[i] })

	v += vParts[len(k)]
	k += secretKeys[0]

	return strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(k+v)), "="), ttl, nil
}

func (mc *MyCrypto) Decrypt(s string) (ret interface{}, err error) {
	if l := len(s) % 4; l > 0 {
		s += strings.Repeat("=", 4-l)
	}

	if b, err := base64.URLEncoding.DecodeString(s); err != nil {
		return ret, err
	} else {
		s = string(b)
	}

	if len(s) < 10*2 {
		return ret, errors.New("unknown")
	}

	k := s[0:10]
	v := s[10:]
	eachPart := int(math.Floor(float64(len(v) / 10)))
	vParts := make([]string, 10)
	for i := 0; i < 9; i++ {
		vParts[i] = v[(i * eachPart) : (i*eachPart)+eachPart]
	}
	vParts[9] = v[9*eachPart:]

	vParts2 := make([]string, 10)
	for i := 0; i < 10; i++ {
		idx := -1
		content := ""
		for idx, content = range mc.encryptMap {
			if strings.Contains(content, k[i:i+1]) {
				break
			}
		}
		vParts2[idx] = vParts[i]
	}

	v = strings.Join(vParts2, "")
	kSalt := v[len(v)-2:]

	if _, err := strconv.Atoi(kSalt); err != nil {
		return ret, errors.New("invalid token")
	}

	lSalt, _ := strconv.Atoi(kSalt[1:2])
	rSalt, _ := strconv.Atoi(kSalt[0:1])

	if ret, err = hex.DecodeString(v[lSalt : len(v)-int(2+rSalt)]); err != nil {
		return ret, err
	}

	var data MyCrypto
	if err = json.Unmarshal(ret.([]byte), &data); err != nil {
		return ret, err
	}

	if data.Expired > 0 && data.Expired <= time.Now().Unix() {
		return data, errors.New("expired")
	}

	return data.Data, nil
}

func DecryptCred(logId string, s string) string {
	mc, err := NewCrypto("")
	if err != nil {
		WriteLog(fmt.Sprintf("%s; DecryptCred; NewCrypto; error: %+v", logId, err), LogLevelError)
		return s
	}

	data, err := mc.Decrypt(s)
	if err != nil {
		WriteLog(fmt.Sprintf("%s; DecryptCred; Decrypt; error: %+v", logId, err), LogLevelError)
		return s
	}
	switch data.(type) {
	case string:
		return data.(string)
	default:
		WriteLog(fmt.Sprintf("%s; DecryptCred; mismatch data type; data: %+v", logId, data), LogLevelInfo)
	}

	return s
}
