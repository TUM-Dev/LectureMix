package main

import (
	"context"
	"time"

	"bitbucket.org/bertimus9/systemstat"
	"k8s.io/klog"
)

type metrics struct {
	compSinkStats srtStats
	pipelineStats pipelineStats // Updated by bus watch on main thread
	cpu           systemstat.CPUSample
	mem           systemstat.MemSample
	loadAvg       systemstat.LoadAvgSample
}

func (d *daemon) metricsProcess(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			cpu := systemstat.GetCPUSample()
			mem := systemstat.GetMemSample()
			loadAvg := systemstat.GetLoadAvgSample()

			srtStats, err := d.srtStatistics()
			if err != nil {
				klog.Warningf("failed to retrieve statistics from srtsinks: %v", err)
				continue
			}

			srtCompStats := srtStats[0]

			d.mu.Lock()
			d.metrics.cpu = cpu
			d.metrics.mem = mem
			d.metrics.loadAvg = loadAvg
			d.metrics.compSinkStats = *srtCompStats
			d.mu.Unlock()

			time.Sleep(time.Second * 1)
		}
	}
}
