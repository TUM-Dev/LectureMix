package main

// TODO: this should be upstreamed

// #cgo pkg-config: glib-2.0 gstreamer-1.0
// #include <glib-object.h>
// #include <gio/gio.h>
// #include <gst/gst.h>
import "C"
import (
	"errors"
	"net"
	"unsafe"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

// GValueArray is sadly not in go-glib and I am to lazy to upstream it
func convertGValueArray(ptr unsafe.Pointer) ([]interface{}, error) {
	valueArray := (*C.GValueArray)(ptr)

	nValues := int(valueArray.n_values)
	values := valueArray.values // Pointer to the array of C.GValue
	cSlice := unsafe.Slice(values, nValues)

	// Map the C.GValue slice to glib.Value slice
	goSlice := make([]interface{}, nValues)
	var err error
	for i, cVal := range cSlice {
		goSlice[i], err = glib.ValueFromNative(unsafe.Pointer(&cVal)).GoValue()
		if err != nil {
			return nil, err
		}
	}

	return goSlice, nil
}

func pipelineGetConfiguredLatency(p *gst.Pipeline) gst.ClockTime {
	ptr := unsafe.Pointer(p.Instance())
	ctime := C.gst_pipeline_get_configured_latency((*C.GstPipeline)(ptr))

	return gst.ClockTime(ctime)
}

const (
	gSocketFamilyInvalid int = 0
	gSocketFamilyUnix    int = 1
	gSocketFamilyIPv4    int = 2
	gSocketFamilyIPv6    int = 10
)

// Check that ptr is an GInetSocketAddress
func inetSocketAddressIP(ptr unsafe.Pointer) (net.IP, uint16, error) {
	// TODO(hugo) check that ptr is actually a GInetSocketAddress instance
	sockAddr := (*C.GInetSocketAddress)(ptr)

	port := uint16(C.g_inet_socket_address_get_port(sockAddr))

	inetAddr := C.g_inet_socket_address_get_address(sockAddr)
	if inetAddr == nil {
		return nil, 0, errors.New("failed to retrieve address from InetSocketAddress instance")
	}

	// Transfer: FULL
	rawAddr := C.g_inet_address_to_string(inetAddr)
	if rawAddr == nil {
		return nil, 0, errors.New("failed to convert GInetAddress to string")
	}
	defer C.free(unsafe.Pointer(rawAddr))

	addr := C.GoString(rawAddr)
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, 0, errors.New("failed to parse ip")
	}

	return ip, port, nil
}
