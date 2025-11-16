package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"regexp"
)

func BsonObjectIDtoString(input interface{}) string {
	re := regexp.MustCompile(`ObjectID\("([0-9a-fA-F]{24})"\)`)
	str := fmt.Sprintf("%v", input)
	match := re.FindStringSubmatch(str)
	return match[0]
}

func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateEventCode generates a unique alphanumeric event code
// Returns a 6-character uppercase code (e.g., "ABC123")
func GenerateEventCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6

	b := make([]byte, codeLength)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range b {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}
