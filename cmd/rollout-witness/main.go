package main

import (
	"fmt"
	"math"
	"net/http"
	"slices"
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
	successfulLatencies := []time.Duration{}

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
			successfulLatencies = append(successfulLatencies, duration)
		}

		fmt.Println("HTTP status:", response.StatusCode)
		response.Body.Close()
	}

	totalDuration := time.Duration(0)

	for _, duration := range latencies {
		totalDuration += duration
	}

	totalFailures := failedRequests + connectionErrors

	averageLatency := time.Duration(0)
	minLatency := time.Duration(0)
	maxLatency := time.Duration(0)
	p95Latency := time.Duration(0)

	if len(latencies) > 0 {
		averageLatency = totalDuration / time.Duration(len(latencies))

		minLatency = latencies[0]
		maxLatency = latencies[0]

		for _, duration := range latencies {
			if duration < minLatency {
				minLatency = duration
			}

			if duration > maxLatency {
				maxLatency = duration
			}
		}

		slices.Sort(latencies)

		p95Rank := math.Ceil(0.95 * float64(len(latencies)))
		p95Index := int(p95Rank) - 1

		p95Latency = latencies[p95Index]
	}

	totalSuccessfulDuration := time.Duration(0)

	for _, duration := range successfulLatencies {
		totalSuccessfulDuration += duration
	}

	averageSuccesfulLatency := time.Duration(0)

	if len(successfulLatencies) > 0 {
		averageSuccesfulLatency = totalSuccessfulDuration / time.Duration(len(successfulLatencies))
	}

	if totalRequests > 0 {
		failureRate = (float64(totalFailures) / float64(totalRequests)) * 100
	}

	fmt.Println(len(latencies))

	fmt.Println("Rollout Witness Summary")
	fmt.Println("Total Requests:", totalRequests)
	fmt.Println("Successful Requests:", successfulRequests)
	fmt.Println("HTTP 5xx failures:", failedRequests)
	fmt.Println("Connection errors:", connectionErrors)
	fmt.Println("Total failures:", totalFailures)
	fmt.Println("Failure Rate:", failureRate, "%")
	fmt.Println("Average latency:", averageLatency)
	fmt.Println("Total successful duration:", totalSuccessfulDuration)
	fmt.Println("Average succesful latency:", averageSuccesfulLatency)
	fmt.Println("Minimum latency:", minLatency)
	fmt.Println("Maximum latency:", maxLatency)
	fmt.Println("P95 latency:", p95Latency)
}
