package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const usage = `Usage of simple-http-st:
  -g, --goroutine-count Number of parallel workers (goroutines). Default: 1
  -d, --duration Test duration. Default 1m
  -H, --headers Headers for http resuest. Default: <nil]>
  -h, --help prints help information
`

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "Headers"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var goroutineCount int
	var headers arrayFlags
	var durationString string
	flag.IntVar(&goroutineCount, "goroutine-count", 1, "")
	flag.IntVar(&goroutineCount, "g", 1, "")
	flag.Var(&headers, "headers", "")
	flag.Var(&headers, "H", "")
	flag.StringVar(&durationString, "duration", "1m", "")
	flag.StringVar(&durationString, "d", "1m", "")

	flag.Usage = func() { fmt.Print(usage) }

	flag.Parse()

	urlString := flag.Args()

	// Check URL
	if len(urlString) != 1 {
		log.Panic("Need to set 1 URL for Stress Test")
	}
	_, err := url.ParseRequestURI(urlString[0])
	if err != nil {
		log.Panic(err)
	}

	// Convert duration from string to time duration
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		log.Panic(err)
	}

	// Parse Headers
	h := map[string]string{}
	for _, v := range headers {
		header := strings.Split(v, ": ")
		if len(header) != 2 {
			log.Panicf("Can't convert %v to header", header)
		}
		h[header[0]] = header[1]
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	testAledyStarted := false
	var requestsCounter int
	var requestsDurations []int64

	for {
		select {
		case <-ctxTimeout.Done():
			durations := requestsDurations
			requestCount := requestsCounter

			sort.SliceStable(durations, func(i, j int) bool {
				return durations[i] < durations[j]
			})

			fmt.Printf("Request Count:\t\t\t%+v\n", requestCount)
			fmt.Printf("Avg:\t\t\t\t%v ms\n", math.Round(getAvg(durations))*100/100)
			startLenfFor50 := int(float32(len(durations)) * 0.5)
			startLenfFor60 := int(float32(len(durations)) * 0.6)
			startLenfFor70 := int(float32(len(durations)) * 0.7)
			startLenfFor80 := int(float32(len(durations)) * 0.8)
			startLenfFor90 := int(float32(len(durations)) * 0.9)
			startLenfFor95 := int(float32(len(durations)) * 0.95)
			startLenfFor99 := int(float32(len(durations)) * 0.99)
			startLenfFor999 := int(float32(len(durations)) * 0.999)

			fmt.Printf("Avg for 50 percentile:\t\t%v ms\n", math.Round(getAvg(durations[startLenfFor50:])*100/100))
			fmt.Printf("Avg for 60 percentile:\t\t%v ms\n", math.Round(getAvg(durations[startLenfFor60:])*100/100))
			fmt.Printf("Avg for 70 percentile:\t\t%v ms\n", math.Round(getAvg(durations[startLenfFor70:])*100/100))
			fmt.Printf("Avg for 80 percentile:\t\t%v ms\n", math.Round(getAvg(durations[startLenfFor80:])*100/100))
			fmt.Printf("Avg for 90 percentile:\t\t%v ms\n", math.Round(getAvg(durations[startLenfFor90:])*100/100))
			fmt.Printf("Avg for 95 percentile:\t\t%v ms\n", math.Round(getAvg(durations[startLenfFor95:])*100/100))
			fmt.Printf("Avg for 99 percentile:\t\t%v ms\n", math.Round(getAvg(durations[startLenfFor99:])*100/100))
			fmt.Printf("Avg for 99.9 percentile:\t%v ms\n", math.Round(getAvg(durations[startLenfFor999:])*100/100))

			return
		default:
			if !testAledyStarted {
				log.Info("Start Stress Test")
				go startTest(urlString[0], h, goroutineCount, &requestsCounter, &requestsDurations)
			}
			testAledyStarted = true
		}
	}

}

func startTest(url string, headers map[string]string, goroutineCount int, requestsCounter *int, requestsDurations *[]int64) {
	limit := make(chan struct{}, goroutineCount)
	for {
		limit <- struct{}{}

		go func() {
			defer func() {
				<-limit
			}()

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Panic(err)
			}

			if len(headers) != 0 {
				for k, v := range headers {
					req.Header.Set(k, v)
				}
			}
			client := &http.Client{}

			start := time.Now()
			resp, err := client.Do(req)
			if err != nil {
				log.Warn(err)
			}
			defer resp.Body.Close()
			elapsed := time.Since(start).Milliseconds()

			*requestsCounter += 1
			*requestsDurations = append(*requestsDurations, elapsed)

		}()
	}
}

func getAvg(slice []int64) float64 {
	var sum int64

	for _, v := range slice {
		sum += v
	}

	return float64(sum) / float64(len(slice))
}
