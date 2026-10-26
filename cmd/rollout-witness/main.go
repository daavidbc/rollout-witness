package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Rollout Witness starting...")

	successfulRequests := 0
	failedRequests := 0
	connectionErrors := 0
	failureRate := 0.0
	totalRequests := 0

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	latencies := []time.Duration{}

	for i := 1; i <= 10; i++ {
		totalRequests++

		fmt.Println("Request number:", i)

		start := time.Now()

		response, err := client.Get("http://podinfo.lab.test/readyz")

		duration := time.Since(start)

		latencies = append(latencies, duration)

		fmt.Println("Duration", duration)

		if err != nil {
			fmt.Println("Request failed:", err)
			connectionErrors++
			continue
		}

		if response.StatusCode >= 500 {
			failedRequests++
		}

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			successfulRequests++
		}

		fmt.Println("HTTP status:", response.StatusCode)
		response.Body.Close()
	}

	totalFailures := failedRequests + connectionErrors

	if totalRequests > 0 {
		failureRate = (float64(totalFailures) / float64(totalRequests)) * 100
	}

	fmt.Println(len(latencies))
	fmt.Println("Rollout Witness Summary")
	fmt.Println("Total Requests:", totalRequests)
	fmt.Println("Successful Requests:", successfulRequests)
	fmt.Println("HTTP 5xx failures:", failedRequests)
	fmt.Println("Connection errors:", connectionErrors)
	fmt.Println("Total failures", totalFailures)
	fmt.Println("Failure Rate:", failureRate, "%")
}
