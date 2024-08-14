package main

import (
	"fmt"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	attacker := vegeta.NewAttacker()
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:8080/orders/a7602a0248524178ac9cdd2b31d6be2d",
	})

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, vegeta.Rate{Freq: 1000000000, Per: time.Second}, 3*time.Second, "GET order test") {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Println("Latency stats")
	fmt.Printf(" - Mean: %s\n", metrics.Latencies.Mean)
	fmt.Printf(" - 99th percentile: %s\n", metrics.Latencies.P99)
	fmt.Printf(" - Max: %s\n", metrics.Latencies.Max)
	fmt.Println("\nOverall")
	fmt.Printf(" - Total success: %.5f\n", metrics.Success)
	fmt.Printf(" - Errors: %v\n", metrics.Errors)
}
