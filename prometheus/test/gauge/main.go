package main

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jkprj/jkfr/prometheus"
	"github.com/jkprj/jkfr/prometheus/gauge"
	"github.com/jkprj/jkfr/prometheus/psutil"
)

var goroutineCount int64 = 0

func TestGetGauge() {
	count := 100
	labels := make(map[string]string)
	labels["statisType"] = "bash_dataproxy_send_sum_1m"
	promGauge := gauge.GetGauge("bash_packets_statis_1m", labels)
	for i := 1; i <= 3; i++ {
		log.Printf("---PromGauge--- count: %d", count)
		promGauge.Set(float64(count))

		time.Sleep(60 * time.Second)
		count += 10
	}

	log.Printf("---------TestGetGauge Over----------")
}

func TestSingalLabelGauge() {
	labels := make(map[string]string)
	labels["statisType"] = "bash_dataproxy_send_sum_1m"
	pUGauge := gauge.GetGauge("bash_packets_statis_1m", labels)

	count1 := 1
	for i := 1; i <= 3; i++ {
		//log.Printf("---x--- count1: %d", count1)
		pUGauge.Set(float64(count1))

		time.Sleep(time.Second)
		count1 += 1
	}

	goroutineCount -= 1
	log.Printf("---------TestSingalLabelGauge finish, goroutineCount: %d----------", goroutineCount)
}

func TestGaugeGoroutineSec() {
	for {
		goroutineCount += 1
		//log.Printf("---------go TestSingalLabelGauge, goroutineCount: %d----------", goroutineCount)
		go TestSingalLabelGauge()
		time.Sleep(3 * time.Second)
	}

	log.Printf("---------TestGaugeGoroutineSec finish----------")
}

func TestMultiLabelsGauge() {

	count2 := int32(1)
	labels := make(map[string]string)
	labels["cpuUsed"] = "greater_than_80_percent"
	labels["memUsed"] = "greater_than_50_percent"
	pUGauge := gauge.GetGauge("process_cpu_mem_check", labels)

	var wg sync.WaitGroup
	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go func(goNum int) {
			defer wg.Done()
			for {
				//log.Printf("---%d--- count2: %d", goNum, count2)
				pUGauge.Set(float64(count2))
				if count2 > 1 {
					atomic.AddInt32(&count2, -1)
				} else {
					atomic.AddInt32(&count2, 1)
				}
				time.Sleep(time.Millisecond * 100)
			}
			//log.Printf("---------goroutine %d exit----------", goNum)
		}(i)
	}
	wg.Wait()

	log.Printf("---------TestMultiLabelsGauge finish----------")
}

func main() {
	go prometheus.Run(":9701")
	//go TestSingalLabelGauge()

	go psutil.ProcCPUAndMEMGauge()

	//TestGaugeGoroutineSec()
	TestMultiLabelsGauge()
	log.Fatal("---------main Over----------")
}
