package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type NumberResponse struct {
	Numbers []int `json:"numbers"`
}

type Response struct {
	Numbers         []int   `json:"numbers"`
	WindowPrevState []int   `json:"windowPrevState"`
	WindowCurrState []int   `json:"windowCurrState"`
	Avg             float64 `json:"avg"`
}

var (
	windowSize     = 10
	numberWindows  = make(map[string][]int)
	windowMutex    sync.Mutex
	client         = &http.Client{Timeout: 500 * time.Millisecond}
	bearerToken    = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJNYXBDbGFpbXMiOnsiZXhwIjoxNzE3MjE4NjM5LCJpYXQiOjE3MTcyMTgzMzksImlzcyI6IkFmZm9yZG1lZCIsImp0aSI6ImFmNWY0ZGQ4LWU0YTEtNDQyMS04NmM5LThmNmQ3NGVmY2IyOCIsInN1YiI6ImdhdXJhdi4yMTI1Y3NlMTE2NkBraWV0LmVkdSJ9LCJjb21wYW55TmFtZSI6IktJRVQgR3JvdXAgb2YgSW5zdGl0dXRpb25zIiwiY2xpZW50SUQiOiJhZjVmNGRkOC1lNGExLTQ0MjEtODZjOS04ZjZkNzRlZmNiMjgiLCJjbGllbnRTZWNyZXQiOiJmcVJRc2ZzWml1SERWc2V0Iiwib3duZXJOYW1lIjoiR2F1cmF2Iiwib3duZXJFbWFpbCI6ImdhdXJhdi4yMTI1Y3NlMTE2NkBraWV0LmVkdSIsInJvbGxObyI6IjIxMDAyOTAxMDAwNjQifQ.sdAY-3KS4EclydWx-0l6HzLO0KuFxl97v1t_mzxELJo"
	testServerURLs = map[string]string{
		"p": "http://20.244.56.144/test/primes",
		"f": "http://20.244.56.144/test/fibo",
		"e": "http://20.244.56.144/test/even",
		"r": "http://20.244.56.144/test/rand",
	}
)

func fetchNumbers(numberType string) ([]int, error) {
	url, ok := testServerURLs[numberType]
	if !ok {
		return nil, fmt.Errorf("invalid number type")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", bearerToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result NumberResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Debug Statement: Print the fetched numbers
	log.Printf("Fetched numbers for type %s: %v\n", numberType, result.Numbers)
	return result.Numbers, nil
}

func calculateAverage(numbers []int) float64 {
	if len(numbers) == 0 {
		return 0
	}
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	return float64(sum) / float64(len(numbers))
}

func handleNumbers(w http.ResponseWriter, r *http.Request) {
	numberType := r.URL.Path[len("/numbers/"):]
	// Debug Statement: Print the request handling info
	log.Printf("Handling request for number type: %s\n", numberType)

	newNumbers, err := fetchNumbers(numberType)
	if err != nil {
		http.Error(w, "Failed to fetch numbers", http.StatusInternalServerError)
		return
	}

	windowMutex.Lock()
	prevState := append([]int(nil), numberWindows[numberType]...)
	currentState := append(numberWindows[numberType], newNumbers...)
	uniqueNumbers := removeDuplicates(currentState)
	if len(uniqueNumbers) > windowSize {
		uniqueNumbers = uniqueNumbers[len(uniqueNumbers)-windowSize:]
	}
	numberWindows[numberType] = uniqueNumbers
	windowMutex.Unlock()

	avg := calculateAverage(uniqueNumbers)

	response := Response{
		Numbers:         newNumbers,
		WindowPrevState: prevState,
		WindowCurrState: uniqueNumbers,
		Avg:             avg,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func removeDuplicates(numbers []int) []int {
	seen := make(map[int]bool)
	result := []int{}
	for _, num := range numbers {
		if !seen[num] {
			seen[num] = true
			result = append(result, num)
		}
	}
	return result
}

func main() {
	http.HandleFunc("/numbers/", handleNumbers)
	log.Println("Server running on port 9876")
	log.Fatal(http.ListenAndServe(":9876", nil))
}
