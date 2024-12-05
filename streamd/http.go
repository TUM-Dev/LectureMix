package main

import (
	"fmt"
	"net/http"

	"github.com/go-gst/go-gst/gst"
)

type httpServer struct {
	daemonController
}

func writeSRTStatsMeta(w http.ResponseWriter) {
	fmt.Fprintln(w, "# HELP srt_callers Current number of subscribers to the SRT stream")
	fmt.Fprintln(w, "# TYPE srt_callers gauge")

	fmt.Fprintln(w, "# HELP srt_send_bytes_total Total bytes sent across all callers")
	fmt.Fprintln(w, "# TYPE srt_send_bytes_total counter")

	fmt.Fprintln(w, "# HELP srt_send_rate Send rate in Mbps")
	fmt.Fprintln(w, "# TYPE srt_send_rate gauge")

	fmt.Fprintln(w, "# HELP srt_bandwidth Bandwidth in Mbps")
	fmt.Fprintln(w, "# TYPE srt_bandwidth gauge")

	fmt.Fprintln(w, "# HELP srt_rtt_seconds RTT in s")
	fmt.Fprintln(w, "# TYPE srt_rtt_seconds gauge")

	fmt.Fprintln(w, "# HELP srt_negotiated_latency_seconds Negotiated latency in s")
	fmt.Fprintln(w, "# TYPE srt_negotiated_latency_seconds gauge")

	fmt.Fprintln(w, "# HELP srt_sent_bytes_total Total bytes sent")
	fmt.Fprintln(w, "# TYPE srt_sent_bytes_total counter")

	fmt.Fprintln(w, "# HELP srt_retransmitted_bytes_total Total bytes retransmitted")
	fmt.Fprintln(w, "# TYPE srt_retransmitted_bytes_total counter")

	fmt.Fprintln(w, "# HELP srt_sent_dropped_bytes_total Total bytes retransmitted")
	fmt.Fprintln(w, "# TYPE srt_sent_dropped_bytes_total counter")

	fmt.Fprintln(w, "# HELP srt_packets_sent_total Total packets sent")
	fmt.Fprintln(w, "# TYPE srt_packets_sent_total counter")

	fmt.Fprintln(w, "# HELP srt_packets_sent_lost_total Total packets lost")
	fmt.Fprintln(w, "# TYPE srt_packets_sent_lost_total counter")

	fmt.Fprintln(w, "# HELP srt_packets_sent_dropped_total Total packets dropped")
	fmt.Fprintln(w, "# TYPE srt_packets_sent_dropped_total counter")

	fmt.Fprintln(w, "# HELP srt_packets_retransmitted_total Total packets retransmitted")
	fmt.Fprintln(w, "# TYPE srt_packets_retransmitted_total counter")

	fmt.Fprintln(w, "# HELP srt_packets_ack_received_total Number of acks received")
	fmt.Fprintln(w, "# TYPE srt_packets_ack_received_total counter")

	fmt.Fprintln(w, "# HELP srt_packets_nack_received_total Number of nacks received")
	fmt.Fprintln(w, "# TYPE srt_packets_nack_received_total counter")
}

func writeSRTStats(w http.ResponseWriter, s *srtStats, sink string) {
	srtTime := s.time.UnixMilli()
	fmt.Fprintf(w, "srt_callers{sink=\"%s\"} %d %d\n", sink, len(s.callers), srtTime)

	// Total bytes sent
	fmt.Fprintf(w, "srt_send_bytes_total{sink=\"%s\"} %d %d\n", sink, s.bytesSendTotal, srtTime)

	// Send rate per caller
	for _, caller := range s.callers {
		common := fmt.Sprintf("address=\"%s\", port=\"%d\", sink=\"%s\"", caller.callerAddress.String(), caller.callerPort, sink)

		// Send Rate
		fmt.Fprintf(w, "srt_send_rate{%s} %f %d\n", common, caller.sendRateMbps, srtTime)

		// Bandwidth
		fmt.Fprintf(w, "srt_bandwidth{%s} %f %d\n", common, caller.bandwidthMbps, srtTime)

		// Round-trip time (RTT)
		fmt.Fprintf(w, "srt_rtt_seconds{%s} %f %d\n", common, caller.rttMS/1000, srtTime)

		// Negotiated Latency
		fmt.Fprintf(w, "srt_negotiated_latency_seconds{%s} %d %d\n", common, caller.negotiatedLatencyMS/1000, srtTime)

		// Bytes sent
		fmt.Fprintf(w, "srt_sent_bytes_total{%s} %d %d\n", common, caller.bytesSent, srtTime)

		// Bytes Retransmitted
		fmt.Fprintf(w, "srt_retransmitted_bytes_total{%s} %d %d\n", common, caller.bytesRetransmitted, srtTime)

		// Bytes Sent Dropped
		fmt.Fprintf(w, "srt_sent_dropped_bytes_total{%s} %d %d\n", common, caller.bytesSentDropped, srtTime)

		// Packets sent
		fmt.Fprintf(w, "srt_packets_sent_total{%s} %d %d\n", common, caller.packetsSent, srtTime)

		// Packets Sent Lost
		fmt.Fprintf(w, "srt_packets_sent_lost_total{%s} %d %d\n", common, caller.packetsSentLost, srtTime)

		// Packets Sent Dropped
		fmt.Fprintf(w, "srt_packets_sent_dropped_total{%s} %d %d\n", common, caller.packetsSentDropped, srtTime)

		// Packets Retransmitted
		fmt.Fprintf(w, "srt_packets_retransmitted_total{%s} %d %d\n", common, caller.packetsRetransmitted, srtTime)

		// Packets Ack Received
		fmt.Fprintf(w, "srt_packets_ack_received_total{%s} %d %d\n", common, caller.packetAckReceived, srtTime)

		// Packets Nack Received
		fmt.Fprintf(w, "srt_packets_nack_received_total{%s} %d %d\n", common, caller.packetNackReceived, srtTime)
	}

}

// Minimalist prometheus exporter
func (h *httpServer) metrics(w http.ResponseWriter, r *http.Request) {
	m := h.metricsSnapshot()

	/* CPU */

	cpuTime := m.cpu.Time.UnixMilli()
	fmt.Fprintf(w, "# HELP linux_proc_user_total Time spent in user mode, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_user_total counter\n")
	fmt.Fprintf(w, "linux_proc_user_total %d %d\n", m.cpu.User, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_system_total Time spent in system mode, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_system_total counter\n")
	fmt.Fprintf(w, "linux_proc_system_total %d %d\n", m.cpu.System, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_iowait_total Time spent waiting for I/O to complete, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_iowait_total counter\n")
	fmt.Fprintf(w, "linux_proc_iowait_total %d %d\n", m.cpu.Iowait, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_irq_total Time spent servicing interrupts, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_irq_total counter\n")
	fmt.Fprintf(w, "linux_proc_irq_total %d %d\n", m.cpu.Irq, cpuTime)

	fmt.Fprintf(w, "# HELP linux_proc_softirq_total Time spent servicing soft interrupts, in ticks\n")
	fmt.Fprintf(w, "# TYPE linux_proc_softirq_total counter\n")
	fmt.Fprintf(w, "linux_proc_softirq_total %d %d\n", m.cpu.SoftIrq, cpuTime)

	/* Memory */

	memTime := m.mem.Time.UnixMilli()
	fmt.Fprintf(w, "# HELP linux_mem_used_bytes Amount of memory used, in bytes\n")
	fmt.Fprintf(w, "# TYPE linux_mem_used_bytes gauge\n")
	fmt.Fprintf(w, "linux_mem_used_bytes %d %d\n", m.mem.MemUsed*1024, memTime)

	fmt.Fprintf(w, "# HELP linux_mem_free_bytes Amount of free memory, in bytes\n")
	fmt.Fprintf(w, "# TYPE linux_mem_free_bytes gauge\n")
	fmt.Fprintf(w, "linux_mem_free_bytes %d %d\n", m.mem.MemFree*1024, memTime)

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

	writeSRTStatsMeta(w)
	writeSRTStats(w, &m.compSinkStats, "combined")
	writeSRTStats(w, &m.presentSinkStats, "present")
	writeSRTStats(w, &m.camSinkStats, "camera")

	/* GStreamer Statistics */

	for k, v := range m.pipelineStats.qosEvents {
		fmt.Fprintf(w, "# HELP gst_qos_events_total Number of qos events\n")
		fmt.Fprintf(w, "# TYPE gst_qos_events_total gauge\n")
		fmt.Fprintf(w, "gst_qos_events_total{source=\"%s\"} %d\n", k, v)
	}
}

const (
	graphDetailMediaType        = "media-type"
	graphDetailCaps             = "caps"
	graphDetailNonDefaultParams = "non-default-params"
	graphDetailStates           = "states"
	graphDetailFullParams       = "full-params"
	graphDetailAll              = "all"
	graphDetailVerbose          = "verbose"
)

func (h *httpServer) graph(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	val := q.Get("details")

	var details gst.DebugGraphDetails
	switch val {
	case graphDetailMediaType:
		details = gst.DebugGraphShowMediaType
	case graphDetailCaps:
		details = gst.DebugGraphShowCapsDetails
	case graphDetailNonDefaultParams:
		details = gst.DebugGraphShowNonDefaultParams
	case graphDetailStates:
		details = gst.DebugGraphShowStates
	case graphDetailFullParams:
		details = gst.DebugGraphShowPullParams
	case graphDetailAll:
		details = gst.DebugGraphShowAll
	case graphDetailVerbose:
		details = gst.DebugGraphShowVerbose
	default:
		details = gst.DebugGraphShowStates
	}

	dot := h.daemonController.graph(details)
	w.Write([]byte(dot))
	w.Header().Add("Content-Type", "text/vnd.graphviz")
}

func (h *httpServer) setupHTTPHandlers() {
	http.HandleFunc("/metrics", h.metrics)
	http.HandleFunc("/graph", h.graph)
}
