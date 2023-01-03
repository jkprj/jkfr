package main

import (
	"jkfr/prometheus"
	"jkfr/prometheus/counter"
	"log"
	"sync"
	"time"
)

func TestGetCounter() {
	count := 100
	labels := make(map[string]string)
	labels["statisType"] = "bash_dataproxy_send_sum_1m"
	promCounter := counter.GetCounter("bash_packets_statis_1m", labels)
	for i := 1; i <= 3; i++ {
		log.Printf("---1--- count: %d", count)
		promCounter.Add(float64(count))

		time.Sleep(60 * time.Second)
		count += 10
	}

	log.Fatal("---------TestGetCounter Over----------")
}

func TestSingalLabelCounter() {

	labels := make(map[string]string)
	labels["serviceLog"] = "consul_error_total"
	pUCounter := counter.GetCounter("log_error_total", labels)

	count := 0
	for i := 1; i <= 60; i++ {
		log.Printf("---2--- count: %d", count)
		pUCounter.Add(float64(count))

		time.Sleep(60 * time.Second)
		count += 10
	}

	log.Fatal("---------TestSingalLabelGauge finish----------")
}

func TestMultiLabelsCounter() {
	labels := make(map[string]string)
	labels["code"] = "200"
	labels["method"] = "GET"
	pUCounter := counter.GetCounter("http_requests_total", labels)

	for i := 1; i <= 60; i++ {
		log.Printf("---Inc---")
		pUCounter.Inc()

		time.Sleep(60 * time.Second)
	}

	log.Printf("---------TestMultiLabelsCounter finish----------")
}

func TestBashStatisCounter() {

	labels := make(map[string]string)
	labels["statisOnNode"] = "dataproxy_send_total"
	pDPCounter := counter.GetCounter("bash_packets_total", labels)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			log.Printf("---dataproxy_send_total---")
			pDPCounter.Add(5)

			time.Sleep(time.Second)
		}
	}()

	labels = make(map[string]string)
	labels["statisOnNode"] = "forwarder_send_total"
	pFWSCounter := counter.GetCounter("bash_packets_total", labels)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			log.Printf("---forwarder_send_total---")
			pFWSCounter.Inc()

			time.Sleep(time.Second)
		}
	}()

	labels = make(map[string]string)
	labels["statisOnNode"] = "forwarder_recv_total"
	pFWRCounter := counter.GetCounter("bash_packets_total", labels)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			log.Printf("---forwarder_recv_total---")
			pFWRCounter.Add(3)

			time.Sleep(time.Second)
		}
	}()

	labels = make(map[string]string)
	labels["statisOnNode"] = "producer_recv_total"
	pPDCounter := counter.GetCounter("bash_packets_total", labels)

	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		sleepCount := 0
		for {
			log.Printf("---producer_recv_total---")
			sleepCount = sleepCount % 60
			if sleepCount == 0 {
				if count == 2 {
					count = 4
				} else {
					count = 2
				}
			}
			pPDCounter.Add(float64(count))

			time.Sleep(time.Second)
			sleepCount++
		}
	}()

	wg.Wait()
	log.Fatal("-----1----Test Counter finish----------")
}

func main() {

	go prometheus.Run(":9700")

	TestBashStatisCounter()
}
