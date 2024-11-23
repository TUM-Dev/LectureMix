package main

// TODO: this should be upstreamed

// #cgo pkg-config: glib-2.0 gstreamer-1.0
// #include <glib-object.h>
// #include <gst/gst.h>
import "C"
import (
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
