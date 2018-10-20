package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strings"
	"time"
)

type stringType int

const (
	numericType stringType = iota
	alphanumericType
	alphabeticalType
)

func toJSON(params map[string][]string) string {
	flattenedParams := make(map[string]string)

	for k, p := range params {
		flattenedParams[k] = p[0]
	}

	js, err := json.Marshal(flattenedParams)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(string(js), ",", ", ", -1)
}

func getRandomASCIIIndex(t stringType) int {
	rand.Seed(time.Now().UnixNano())
	switch t {
	case numericType:
		return randInt(48, 57)
	case alphanumericType:
		coinFlip := rand.Intn(2)
		if coinFlip == 0 {
			// Random character is a number
			return randInt(48, 57)
		}
		// Random character is a letter
		return randInt(65, 90)
	case alphabeticalType:
		coinFlip := rand.Intn(2)
		if coinFlip == 0 {
			// Random uppercase character
			return randInt(65, 90)
		}
		// Random lowercase character
		return randInt(97, 122)
	default:
		return randInt(48, 57)
	}
}

func randomString(l int, t stringType) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(getRandomASCIIIndex(t))
	}

	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
