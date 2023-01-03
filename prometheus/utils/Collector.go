package utils

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type NewHandle func(name string, labelNames []string) prometheus.Collector

type UCollector struct {
	mtCollector    sync.RWMutex
	mtNewCollector sync.Mutex
	mapCollector   map[string]prometheus.Collector

	newHandle NewHandle
}

func NewUCollector(newHandle NewHandle) *UCollector {
	c := new(UCollector)
	c.mapCollector = make(map[string]prometheus.Collector)
	c.newHandle = newHandle

	return c
}

func (uc *UCollector) getCollector(key string) prometheus.Collector {
	uc.mtCollector.RLock()
	defer uc.mtCollector.RUnlock()

	if c, ok := uc.mapCollector[key]; ok {
		return c
	}

	return nil
}

func (uc *UCollector) pushCollector(key string, c prometheus.Collector) {
	uc.mtCollector.Lock()
	defer uc.mtCollector.Unlock()

	uc.mapCollector[key] = c
}

func (uc *UCollector) newCollector(name string, labelNames []string) prometheus.Collector {

	c := uc.newHandle(name, labelNames)
	prometheus.MustRegister(c)

	uc.pushCollector(MarshalKey(name, labelNames), c)

	return c
}

func (uc *UCollector) GetCollector(name string, labelNames []string) prometheus.Collector {

	key := MarshalKey(name, labelNames)

	c := uc.getCollector(key)
	if nil != c {
		return c
	}

	uc.mtNewCollector.Lock()
	defer uc.mtNewCollector.Unlock()

	// 再获取一次，防止重复创建
	c = uc.getCollector(key)
	if nil != c {
		return c
	}

	return uc.newCollector(name, labelNames)
}
