package main

import (
	"fmt"
	"net/http"
)

type httpServer struct {
	daemonController
}

// Minimalist prometheus exporter
func (h *httpServer) metrics(w http.ResponseWriter, r *http.Request) {
	m := h.metricsSnapshot()

	/* CPU */

	cpuTime := m.cpu.Time.UnixMilli()
	fmt.Fprintf(w, "# HELP linux_proc_user Time spent in user mode, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_user gauge\n")
	fmt.Fprintf(w, "linux_proc_user %d %d\n", m.cpu.User, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_system Time spent in system mode, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_system gauge\n")
	fmt.Fprintf(w, "linux_proc_system %d %d\n", m.cpu.System, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_iowait Time spent waiting for I/O to complete, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_iowait gauge\n")
	fmt.Fprintf(w, "linux_proc_iowait %d %d\n", m.cpu.Iowait, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_irq Time spent servicing interrupts, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_irq gauge\n")
	fmt.Fprintf(w, "linux_proc_irq %d %d\n", m.cpu.Irq, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_softirq Time spent servicing soft interrupts, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_softirq gauge\n")
	fmt.Fprintf(w, "linux_proc_softirq %d %d\n", m.cpu.SoftIrq, cpuTime)

	/* Memory */

	memTime := m.mem.Time.UnixMilli()
	fmt.Fprintf(w, "# HELP linux_mem_used Amount of memory used, in kB\n")
	fmt.Fprintf(w, "# TYPE linux_mem_used gauge\n")
	fmt.Fprintf(w, "linux_mem_used %d %d\n", m.mem.MemUsed, memTime)

	fmt.Fprintf(w, "# HELP linux_mem_free Amount of free memory, in kB\n")
	fmt.Fprintf(w, "# TYPE linux_mem_free gauge\n")
	fmt.Fprintf(w, "linux_mem_free %d %d\n", m.mem.MemFree, memTime)

	/* TODO(hugo) I/O Status metrics via `iostat` when recording is implemented */

	/* Load Average */

	// The timestamp is an int64 (milliseconds since epoch, i.e. 1970-01-01
	// 00:00:00 UTC, excluding leap seconds), represented as required by Go's
	// ParseInt() function.
	loadAvgTime := m.loadAvg.Time.UnixMilli()
	fmt.Fprintf(w, "# HELP load_avg_one Load average over one minute\n")
	fmt.Fprintf(w, "# TYPE load_avg_one gauge\n")
	fmt.Fprintf(w, "load_avg_one %f %d\n", m.loadAvg.One, loadAvgTime)

	fmt.Fprintf(w, "# HELP load_avg_five Load average over five minutes\n")
	fmt.Fprintf(w, "# TYPE load_avg_five gauge\n")
	fmt.Fprintf(w, "load_avg_five %f %d\n", m.loadAvg.Five, loadAvgTime)

	fmt.Fprintf(w, "# HELP load_avg_fifteen Load average over fifteen minutes\n")
	fmt.Fprintf(w, "# TYPE load_avg_fifteen gauge\n")
	fmt.Fprintf(w, "load_avg_fifteen %f %d\n", m.loadAvg.Fifteen, loadAvgTime)

	/* SRT Statistics */

	srtTime := m.compSinkStats.time.UnixMilli()
	fmt.Fprintf(w, "# HELP srt_callers Current number of subscribers to the SRT stream\n")
	fmt.Fprintf(w, "# TYPE srt_callers gauge\n")
	fmt.Fprintf(w, "srt_callers %d %d\n", len(m.compSinkStats.callers), srtTime)

	// Total bytes sent
	fmt.Fprintf(w, "# HELP srt_bytes_send_total Total bytes sent across all callers\n")
	fmt.Fprintf(w, "# TYPE srt_bytes_send_total counter\n")
	fmt.Fprintf(w, "srt_bytes_send_total %d %d\n", m.compSinkStats.bytesSendTotal, srtTime)

	// Send rate per caller
	// TODO: caller should be identified by IP, but that field is NULL in the srt stats structure
	for i, caller := range m.compSinkStats.callers {
		common := fmt.Sprintf("caller=\"%d\"", i)

		// Send Rate
		fmt.Fprintf(w, "# HELP srt_send_rate Send rate in Mbps")
		fmt.Fprintf(w, "# TYPE srt_send_rate gauge\n")
		fmt.Fprintf(w, "srt_send_rate{%s} %f %d\n", common, caller.sendRateMbps, srtTime)

		// Bandwidth
		fmt.Fprintf(w, "# HELP srt_bandwidth Bandwidth in Mbps")
		fmt.Fprintf(w, "# TYPE srt_bandwidth gauge\n")
		fmt.Fprintf(w, "srt_bandwidth{%s} %f %d\n", common, caller.bandwidthMbps, srtTime)

		// Round-trip time (RTT)
		fmt.Fprintf(w, "# HELP srt_rtt RTT in ms")
		fmt.Fprintf(w, "# TYPE srt_rtt gauge\n")
		fmt.Fprintf(w, "srt_rtt{%s} %f %d\n", common, caller.rttMS, srtTime)

		// Negotiated Latency
		fmt.Fprintf(w, "# HELP srt_negotiated_latency Negotiated latency in ms")
		fmt.Fprintf(w, "# TYPE srt_negotiated_latency gauge\n")
		fmt.Fprintf(w, "srt_negotiated_latency{%s} %d %d\n", common, caller.negotiatedLatencyMS, srtTime)

		// Bytes sent
		fmt.Fprintf(w, "# HELP srt_bytes_sent Total bytes sent")
		fmt.Fprintf(w, "# TYPE srt_bytes_sent gauge\n")
		fmt.Fprintf(w, "srt_bytes_sent{%s} %d %d\n", common, caller.bytesSent, srtTime)

		// Bytes Retransmitted
		fmt.Fprintf(w, "# HELP srt_bytes_retransmitted Total bytes retransmitted")
		fmt.Fprintf(w, "# TYPE srt_bytes_retransmitted gauge\n")
		fmt.Fprintf(w, "srt_bytes_retransmitted{%s} %d %d\n", common, caller.bytesRetransmitted, srtTime)

		// Bytes Sent Dropped
		fmt.Fprintf(w, "# HELP srt_bytes_send_dropped Total bytes retransmitted")
		fmt.Fprintf(w, "# TYPE srt_bytes_send_dropped gauge\n")
		fmt.Fprintf(w, "srt_bytes_send_dropped{%s} %d %d\n", common, caller.bytesSentDropped, srtTime)

		// Packets sent
		fmt.Fprintf(w, "# HELP srt_packets_sent Total packets sent")
		fmt.Fprintf(w, "# TYPE srt_packets_sent gauge\n")
		fmt.Fprintf(w, "srt_packets_sent{%s} %d %d\n", common, caller.packetsSent, srtTime)

		// Packets Sent Lost
		fmt.Fprintf(w, "# HELP srt_packets_sent_lost Total packets lost")
		fmt.Fprintf(w, "# TYPE srt_packets_sent_lost gauge\n")
		fmt.Fprintf(w, "srt_packets_sent_lost{%s} %d %d\n", common, caller.packetsSentLost, srtTime)

		// Packets Sent Dropped
		fmt.Fprintf(w, "# HELP srt_packets_sent_dropped Total packets dropped")
		fmt.Fprintf(w, "# TYPE srt_packets_sent_dropped gauge\n")
		fmt.Fprintf(w, "srt_packets_sent_dropped{%s} %d %d\n", common, caller.packetsSentDropped, srtTime)

		// Packets Retransmitted
		fmt.Fprintf(w, "# HELP srt_packets_retransmitted Total packets retransmitted")
		fmt.Fprintf(w, "# TYPE srt_packets_retransmitted gauge\n")
		fmt.Fprintf(w, "srt_packets_retransmitted{%s} %d %d\n", common, caller.packetsRetransmitted, srtTime)

		// Packets Ack Received
		fmt.Fprintf(w, "# HELP srt_packets_ack_received Number of acks received")
		fmt.Fprintf(w, "# TYPE srt_packets_ack_received gauge\n")
		fmt.Fprintf(w, "srt_packets_ack_received{%s} %d %d\n", common, caller.packetAckReceived, srtTime)

		// Packets Nack Received
		fmt.Fprintf(w, "# HELP srt_packets_nack_received Number of nacks received")
		fmt.Fprintf(w, "# TYPE srt_packets_nack_received gauge\n")
		fmt.Fprintf(w, "srt_packets_nack_received{%s} %d %d\n", common, caller.packetNackReceived, srtTime)

		// TODO(hugo): Add receive metrics from 'srtCallerStats'?
	}
}

func (h *httpServer) setupHTTPHandlers() {
	http.HandleFunc("/metrics", h.metrics)
}
