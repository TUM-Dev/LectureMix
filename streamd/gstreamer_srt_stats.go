package main

import (
	"errors"
	"fmt"
	"net"
	"time"
	"unsafe"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

const (
	srtStatsMimetype = "application/x-srt-statistics"
)

// Here is an example of the GstStructure:
//
// application/x-srt-statistics, callers=(GValueArray)<
// "application/x-srt-statistics\,\ packets-sent\=\(gint64\)134\,\
// packets-sent-lost\=\(int\)0\,\ packets-retransmitted\=\(int\)0\,\
// packet-ack-received\=\(int\)0\,\ packet-nack-received\=\(int\)0\,\
// send-duration-us\=\(guint64\)0\,\ bytes-sent\=\(guint64\)31088\,\
// bytes-retransmitted\=\(guint64\)0\,\ bytes-sent-dropped\=\(guint64\)0\,\
// packets-sent-dropped\=\(int\)0\,\
// send-rate-mbps\=\(double\)0.27281692381103867\,\
// negotiated-latency-ms\=\(int\)125\,\ packets-received\=\(gint64\)0\,\
// packets-received-lost\=\(int\)0\,\
// packets-received-retransmitted\=\(int\)0\,\
// packets-received-dropped\=\(int\)0\,\ packet-ack-sent\=\(int\)0\,\
// packet-nack-sent\=\(int\)0\,\ bytes-received\=\(guint64\)0\,\
// bytes-received-lost\=\(guint64\)0\,\ receive-rate-mbps\=\(double\)0\,\
// bandwidth-mbps\=\(double\)12\,\ rtt-ms\=\(double\)100\,\
// caller-address\=\(GSocketAddress\)NULL\;" >, bytes-sent-total=(guint64)25192;

// An srtStats struct captures application/x-srt-statistics data
type srtStats struct {
	callers        []srtCallerStats
	bytesSendTotal uint64

	time time.Time
}

type srtCallerStats struct {
	sendDurationUs  uint64
	sendRateMbps    float64
	receiveRateMbps float64
	bandwidthMbps   float64
	rttMS           float64

	bytesSent          uint64
	bytesRetransmitted uint64
	bytesSentDropped   uint64
	bytesReceived      uint64
	bytesReceivedLost  uint64

	packetsSent          int64 // Don't ask, this is what GStreamer gives me. I want -1 packets please.
	packetsReceived      int64
	packetsSentLost      int
	packetsSentDropped   int
	packetsRetransmitted int
	packetAckReceived    int
	packetNackReceived   int

	packetsReceivedLost          int
	packetsReceivedRetransmitted int
	packetReceivedDropped        int
	packetAckSent                int
	packetNackSent               int

	negotiatedLatencyMS int

	callerAddress net.IP
	callerPort    uint16
}

// Retrieve value from name and convert it to the correct time (known at compile time)
func valueTo[T any](s *gst.Structure, name string, dest *T) error {
	obj, err := s.GetValue(name)
	if err != nil {
		return fmt.Errorf("failed to retrieve '%s': %w", name, err)
	}
	value, ok := obj.(T)
	if !ok {
		return fmt.Errorf("failed to cast value of '%s' to %T", name, *new(T))
	}

	*dest = value
	return nil
}

func valuesTo[T any](s *gst.Structure, props *[]struct {
	dest *T
	name string
}) error {
	for _, prop := range *props {
		err := valueTo(s, prop.name, prop.dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func newSRTStatsFromStructure(s *gst.Structure) (*srtStats, error) {
	stats := &srtStats{}

	stats.time = time.Now()

	mimetype := s.Name()
	if mimetype != srtStatsMimetype {
		return nil, fmt.Errorf("struct has wrong mimetype. Expected '%s' but got '%s'", srtStatsMimetype, mimetype)
	}

	err := valueTo(s, "bytes-sent-total", &stats.bytesSendTotal)
	if err != nil {
		return nil, err
	}

	var ptr unsafe.Pointer
	err = valueTo(s, "callers", &ptr)
	if err != nil {
		// May be absent
		return stats, nil
	}

	arr, err := convertGValueArray(ptr)
	if err != nil {
		return nil, err
	}

	err = stats.convertCallerStats(arr)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *srtStats) convertCallerStats(arr []interface{}) error {
	for idx, entry := range arr {
		sc := srtCallerStats{}
		gs, ok := entry.(*gst.Structure)
		if ok != true {
			return fmt.Errorf("failed to convert GstStructure at %d", idx)
		}

		name := gs.Name()
		if name != srtStatsMimetype {
			return fmt.Errorf("struct has wrong mimetype '%s'", name)
		}

		socketAddress, err := gs.GetValue("caller-address")
		if err != nil {
			return err
		}
		socketAddressObj, ok := socketAddress.(*glib.Object)
		if ok != true {
			return errors.New("caller-address is not a glib object")
		}
		sc.callerAddress, sc.callerPort, err = inetSocketAddressIP(socketAddressObj.Unsafe())
		if err != nil {
			return err
		}

		intProps := []struct {
			dest *int
			name string
		}{
			{&sc.packetsSentLost, "packets-sent-lost"},
			{&sc.packetsSentDropped, "packets-sent-dropped"},
			{&sc.packetsRetransmitted, "packets-retransmitted"},
			{&sc.packetAckReceived, "packet-ack-received"},
			{&sc.packetNackReceived, "packet-nack-received"},
			{&sc.packetsReceivedLost, "packets-received-lost"},
			{&sc.packetsReceivedRetransmitted, "packets-received-retransmitted"},
			{&sc.packetReceivedDropped, "packets-received-dropped"},
			{&sc.packetAckSent, "packet-ack-sent"},
			{&sc.packetNackSent, "packet-nack-sent"},
			{&sc.negotiatedLatencyMS, "negotiated-latency-ms"},
		}
		if err := valuesTo(gs, &intProps); err != nil {
			return err
		}

		uint64Props := []struct {
			dest *uint64
			name string
		}{
			{&sc.sendDurationUs, "send-duration-us"},
			{&sc.bytesSent, "bytes-sent"},
			{&sc.bytesRetransmitted, "bytes-retransmitted"},
			{&sc.bytesSentDropped, "bytes-sent-dropped"},
			{&sc.bytesReceived, "bytes-received"},
			{&sc.bytesReceivedLost, "bytes-received-lost"},
		}
		if err := valuesTo(gs, &uint64Props); err != nil {
			return err
		}

		if err := valueTo(gs, "packets-sent", &sc.packetsSent); err != nil {
			return err
		}
		if err := valueTo(gs, "packets-received", &sc.packetsReceived); err != nil {
			return err
		}

		float64Props := []struct {
			dest *float64
			name string
		}{
			{&sc.sendRateMbps, "send-rate-mbps"},
			{&sc.receiveRateMbps, "receive-rate-mbps"},
			{&sc.bandwidthMbps, "bandwidth-mbps"},
			{&sc.rttMS, "rtt-ms"},
		}
		if err := valuesTo(gs, &float64Props); err != nil {
			return err
		}

		s.callers = append(s.callers, sc)
	}

	return nil
}
