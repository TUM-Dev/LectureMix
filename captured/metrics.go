package main

import (
	"context"
	"time"

	"bitbucket.org/bertimus9/systemstat"
)

type metrics struct {
	compSinkStats srtStats
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

			/*
				srtCompStats, err := d.srtCompSinkStats()
				if err != nil {
					klog.Warningf("failed to retrieve statistics from srtsink serving compositor stream: %v", err)
				}
			*/

			d.mu.Lock()
			d.metrics.cpu = cpu
			d.metrics.mem = mem
			d.metrics.loadAvg = loadAvg
			//d.metrics.compSinkStats = *srtCompStats
			d.mu.Unlock()

			time.Sleep(time.Second * 1)
		}
	}
}
