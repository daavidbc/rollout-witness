package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"slices"
	"time"
)

type RequestResult struct {
	Timestamp  time.Time
	StatusCode int
	Duration   time.Duration
	Error      string
}

func main() {
	fmt.Println("Rollout Witness starting...")

	successfulRequests := 0
	failedRequests := 0
	connectionErrors := 0
	failureRate := 0.0
	totalRequests := 0

	url := flag.String(
		"url",
		"http://podinfo.lab.test/readyz",
		"HTTP endpoint to monitor",
	)

	requests := flag.Int(
		"requests",
		10,
		"Number of HTTP requests",
	)

	timeout := flag.Duration(
		"timeout",
		3*time.Second,
		"HTTP Request timeout",
	)

	interval := flag.Duration(
		"interval",
		1*time.Second,
		"Delay between HTTP requests",
	)

	hostHeader := flag.String(
		"host-header",
		"",
		"HTTP Host header for Ingress routing",
	)

	flag.Parse()

	if *timeout <= 0 {
		fmt.Println("Error: timeout must be greater than zero")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: *timeout,
	}

	latencies := []time.Duration{}
	successfulLatencies := []time.Duration{}
	results := []RequestResult{}

	if *requests <= 0 {
		fmt.Println("Error: requests must be greater than zero")
		os.Exit(1)
	}

	if *interval < 0 {
		fmt.Println("Error: interval must not be negative")
		os.Exit(1)
	}

	for i := 1; i <= *requests; i++ {

		if i > 1 {
			time.Sleep(*interval)
		}

		totalRequests++

		fmt.Println("Request number:", i)

		req, err := http.NewRequest("GET", *url, nil)

		if err != nil {
			fmt.Println("Request creation failed:", err)
			os.Exit(1)
		}

		if *hostHeader != "" {
			req.Host = *hostHeader
		}

		start := time.Now()
		result := RequestResult{
			Timestamp: start,
		}

		response, err := client.Do(req)
		duration := time.Since(start)

		result.Duration = duration

		result = RequestResult{
			Timestamp: time.Now(),
		}

		fmt.Println(result.Timestamp)

		latencies = append(latencies, duration)
		fmt.Println("Duration", duration)

		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			fmt.Println("HTTP request failed:", err)
			connectionErrors++
			continue
		}

		result.StatusCode = response.StatusCode
		results = append(results, result)

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			successfulRequests++
			successfulLatencies = append(successfulLatencies, duration)
		} else {
			failedRequests++
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

	if totalRequests > 0 {
		failureRate = (float64(totalFailures) / float64(totalRequests)) * 100
	}

	fmt.Println("Rollout Witness Summary")
	fmt.Println("Total Requests:", totalRequests)
	fmt.Println("Successful Requests:", successfulRequests)
	fmt.Println("HTTP failures:", failedRequests)
	fmt.Println("Connection errors:", connectionErrors)
	fmt.Println("Total failures:", totalFailures)
	fmt.Println("Failure Rate:", failureRate, "%")
	fmt.Println("Average latency:", averageLatency)
	fmt.Println("Total successful duration:", totalSuccessfulDuration)

	if len(successfulLatencies) > 0 {
		averageSuccessfulLatency := totalSuccessfulDuration / time.Duration(len(successfulLatencies))
		fmt.Println("Average successful latency:", averageSuccessfulLatency)
	} else {
		fmt.Println("Average successful latency: N/A")
	}

	fmt.Println("Minimum latency:", minLatency)
	fmt.Println("Maximum latency:", maxLatency)
	fmt.Println("P95 latency:", p95Latency)

	if totalFailures > 0 {
		os.Exit(1)
	}
}
