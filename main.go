package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Response struct {
	Result int `json:"result"`
}

var cache = make(map[string]int)
var cacheMutex = &sync.Mutex{}

const maxQueryLength = 200

func main() {
	http.HandleFunc("/api/arithmetic", arithmeticHandler)
	fmt.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func validateAccess(r *http.Request) bool {
	return r.Header.Get("User-Access") == "superuser"
}

func validateQuery(query string) bool {
	for _, char := range query {
		if !((char >= '0' && char <= '9') || char == ' ' || char == '+' || char == '-') {
			return false
		}
	}
	return true
}

func parseExpression(expr string) (int, error) {
	cacheMutex.Lock()
	result, exists := cache[expr]
	cacheMutex.Unlock()
	if exists {
		return result, nil
	}

	expr = strings.ReplaceAll(expr, " ", "+")
	currentNum := 0
	sign := 1
	result = 0

	for _, char := range expr {
		if char >= '0' && char <= '9' {
			currentNum = currentNum*10 + int(char-'0')
		} else if char == '+' || char == '-' {
			result += sign * currentNum
			currentNum = 0

			if char == '+' {
				sign = 1
			} else {
				sign = -1
			}
		} else {
			return 0, fmt.Errorf("invalid character: %c", char)
		}
	}

	result += sign * currentNum

	cacheMutex.Lock()
	cache[expr] = result
	cacheMutex.Unlock()

	return result, nil
}

func arithmeticHandler(w http.ResponseWriter, r *http.Request) {
	if !validateAccess(r) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Access denied")
		fmt.Println("Access denied for request:", r.URL.Path)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing query parameter")
		return
	}

	if len(query) > maxQueryLength {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Query parameter too long")
		return
	}

	if !validateQuery(query) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid query parameter")
		return
	}

	result, err := parseExpression(query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := Response{Result: result}
	json.NewEncoder(w).Encode(response)
}
