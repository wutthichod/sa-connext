package utils

import (
	"fmt"
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
