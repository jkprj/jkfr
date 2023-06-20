package psutil

import (
	"log"
	"os"

	"github.com/jkprj/jkfr/prometheus/gauge"

	"github.com/shirou/gopsutil/v3/process"
)

func GetProcess() (process.Process, error) {
	checkPid := os.Getpid()
	ret, err := process.NewProcess(int32(checkPid))
	return *ret, err
}

func GetProcCPUAndMEM(p process.Process) (float64, float32, error) {
	//计算最近一个时间间隔里的CPU使用率
	cpuPer3s, err := p.Percent(3 * 1000 * 1000 * 1000) // 3s
	if err != nil {
		return 0, 0, err
	}

	//计算进程启动后的CPU使用率
	//cpuPerSince, err := p.CPUPercent()

	memPer, err := p.MemoryPercent()
	if err != nil {
		return 0, 0, err
	}

	return cpuPer3s, memPer, nil
}

func ProcCPUAndMEMGauge() {
	p, err := GetProcess()
	if err != nil {
		log.Printf("GetProcess failed")
		return
	}

	procName, err := p.Name()
	pCPUGauge := gauge.GetGauge("process_cpu_mem_usage", map[string]string{procName: "cpuUsed"})
	pMEMGauge := gauge.GetGauge("process_cpu_mem_usage", map[string]string{procName: "memUsed"})

	for {
		cpuPer3s, memPer, err := GetProcCPUAndMEM(p)
		if err != nil {
			log.Printf("GetProcCPUAndMEM failed")
			continue
		}

		pCPUGauge.Set(float64(cpuPer3s))
		pMEMGauge.Set(float64(memPer))
	}
}
