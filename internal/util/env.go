package util

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
)

func GetEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return b
}

func GetEnvIntSlice(key string, defaultValue []int) []int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	values := strings.Split(value, ",")
	intValues := make([]int, len(values))
	for i, value := range values {
		intValues[i], _ = strconv.Atoi(value)
	}
	return intValues
}

func GetEnvStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	values := strings.Split(value, ",")
	stringValues := make([]string, len(values))
	for i, v := range values {
		stringValues[i] = strings.TrimSpace(v)
	}
	return stringValues
}

func GetEnvJSON(key string) map[string]interface{} {
	jsonData := os.Getenv(key)
	if jsonData == "" {
		return nil
	}
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		log.Println(err.Error())
	}
	return result
}

func GetEnvStringIntMap(key string, defaultValue map[string]int) map[string]int {
	mapData := GetEnvJSON(key)
	if mapData == nil {
		return defaultValue
	}
	result := make(map[string]int)
	for k, v := range mapData {
		result[k] = int(v.(float64))
	}
	return result
}
