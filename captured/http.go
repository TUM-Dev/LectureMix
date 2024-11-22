package main

import (
	"fmt"
	"net/http"
)

type httpServer struct {
	daemonController
}

func (h *httpServer) metrics(w http.ResponseWriter, r *http.Request) {
	m := h.metricsSnapshot()

	/* CPU */

	fmt.Fprintf(w, "# HELP linux_proc_user Time spent in user mode, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_user gauge\n")
	fmt.Fprintf(w, "linux_proc_user %f\n", float64(m.cpu.User))

	fmt.Fprintf(w, "# HELP linux_proc_system Time spent in system mode, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_system gauge\n")
	fmt.Fprintf(w, "linux_proc_system %f\n", float64(m.cpu.System))

	fmt.Fprintf(w, "# HELP linux_proc_iowait Time spent waiting for I/O to complete, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_iowait gauge\n")
	fmt.Fprintf(w, "linux_proc_iowait %f\n", float64(m.cpu.Iowait))

	fmt.Fprintf(w, "# HELP linux_proc_irq Time spent servicing interrupts, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_irq gauge\n")
	fmt.Fprintf(w, "linux_proc_irq %f\n", float64(m.cpu.Irq))

	fmt.Fprintf(w, "# HELP linux_proc_softirq Time spent servicing soft interrupts, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_softirq gauge\n")
	fmt.Fprintf(w, "linux_proc_softirq %f\n", float64(m.cpu.SoftIrq))

	/* Memory */

	fmt.Fprintf(w, "# HELP linux_mem_used Amount of memory used, in kB\n")
	fmt.Fprintf(w, "# TYPE linux_mem_used gauge\n")
	fmt.Fprintf(w, "linux_mem_used %f\n", float64(m.mem.MemUsed))

	fmt.Fprintf(w, "# HELP linux_mem_free Amount of free memory, in kB\n")
	fmt.Fprintf(w, "# TYPE linux_mem_free gauge\n")
	fmt.Fprintf(w, "linux_mem_free %f\n", float64(m.mem.MemFree))

	/* Load Average */

	loadAvgTime := m.loadAvg.Time.Unix()
	fmt.Fprintf(w, "# HELP load_avg_one Load average over one minute\n")
	fmt.Fprintf(w, "# TYPE load_avg_one gauge\n")
	fmt.Fprintf(w, "load_avg_one %f %d\n", m.loadAvg.One, loadAvgTime)

	fmt.Fprintf(w, "# HELP load_avg_five Load average over five minutes\n")
	fmt.Fprintf(w, "# TYPE load_avg_five gauge\n")
	fmt.Fprintf(w, "load_avg_five %f %d\n", m.loadAvg.Five, loadAvgTime)

	fmt.Fprintf(w, "# HELP load_avg_fifteen Load average over fifteen minutes\n")
	fmt.Fprintf(w, "# TYPE load_avg_fifteen gauge\n")
	fmt.Fprintf(w, "load_avg_fifteen %f %d\n", m.loadAvg.Fifteen, loadAvgTime)

}

func (h *httpServer) setupHTTPHandlers() {
	http.HandleFunc("/metrics", h.metrics)
}
