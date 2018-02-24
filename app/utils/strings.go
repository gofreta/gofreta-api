package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// MD5 hashes using md5 algorithm
func MD5(str string) string {
	hash := md5.Sum([]byte(str))

	return hex.EncodeToString(hash[:])
}

// UcFirst converts the first character of a string into uppercase.
func UcFirst(str string) string {
	if str == "" {
		return ""
	}

	s := []rune(str)

	return string(unicode.ToUpper(s[0])) + string(s[1:])
}

// Sentenize converts and normalizes string into a sentence.
func Sentenize(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}

	s := []rune(str)
	sentence := string(unicode.ToUpper(s[0])) + string(s[1:])

	lastChar := string(s[len(s)-1:])
	if lastChar != "." && lastChar != "?" && lastChar != "!" {
		return sentence + "."
	}

	return sentence
}

// Random generates and returns random string.
func Random(length int) string {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphabetLength := len(alphabet)

	result := make([]byte, length)
	for i := range result {
		result[i] = alphabet[rand.Intn(alphabetLength)]
	}

	return string(result)
}

// StringInSlice checks whether a string exist in array/slice.
func StringInSlice(str string, list []string) bool {
	if list != nil {
		for _, v := range list {
			if v == str {
				return true
			}
		}
	}

	return false
}
