package gstreamer

// #cgo pkg-config: gstreamer-1.0 gstreamer-base-1.0
// #include <gst/gst.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

func init() {
	C.gst_init(nil, nil)
}

type ParseErrorCode int

// Maps to GstParseError enumeration
const (
	ParseErrorSyntax              ParseErrorCode = 0
	ParseErrorNoSuchElement       ParseErrorCode = 1
	ParseErrorNoSuchProperty      ParseErrorCode = 2
	ParseErrorLink                ParseErrorCode = 3
	ParseErrorCouldNotSetProperty ParseErrorCode = 4
	ParseErrorEmptyBin            ParseErrorCode = 5
	ParseErrorEmpty               ParseErrorCode = 6
	ParseErrorDelayedLink         ParseErrorCode = 7
)

type ParseError struct {
	Code    ParseErrorCode
	Message string
}

func newParseError(gerr *C.struct__GError) ParseError {
	var err ParseError

	if gerr == nil {
		panic("gerr must be a valid pointer to GError")
	}

	err.Code = ParseErrorCode(gerr.code)
	err.Message = C.GoString(gerr.message)

	return err
}

func (p ParseError) Error() string {
	return fmt.Sprintf("(%d): %s", p.Code, p.Message)
}

type Pipeline struct {
	handle    *C.GstElement
	busHandle *C.GstBus
	valid     bool
}

func NewPipeline(description string) (*Pipeline, error) {
	var pipeline Pipeline
	var err error

	// Create a cstring from the pipeline description.  gst_parse_launch
	// duplicates the string, so we can safely free it at the end of this
	// constructor.
	// Transfer: Full
	cs := C.CString(description)
	defer C.free(unsafe.Pointer(cs))

	var gError *C.struct__GError
	// Transfer: Floating
	pipeline.handle = C.gst_parse_launch(cs, &gError)
	if gError != nil {
		defer C.free(unsafe.Pointer(gError))
		return nil, newParseError(gError)
	}
	// Transfer: Full
	pipeline.busHandle = C.gst_element_get_bus(pipeline.handle)

	pipeline.valid = true

	return &pipeline, err
}

func (p *Pipeline) Start() error {
	if !p.valid {
		return errors.New("pipeline object is not valid")
	}

	code := C.gst_element_set_state(p.handle, C.GstState(StatePlaying))
	if StateChangeReturn(code) == StateChangeReturnFailure {
		return errors.New("failed to set pipeline state to playing")
	}

	return nil
}

func (p *Pipeline) Stop() {
	if !p.valid {
		return
	}

	_ = C.gst_element_set_state(p.handle, C.GstState(StateNull))
}

func (p *Pipeline) Unref() {
	if p.valid {
		C.gst_object_unref(C.gpointer(unsafe.Pointer(p.handle)))
		C.gst_object_unref(C.gpointer(unsafe.Pointer(p.busHandle)))
		p.valid = false
	}
}

func (p *Pipeline) String() string {
	return fmt.Sprintf("underlying=(%p)", p.handle)
}
