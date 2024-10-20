package gstreamer

type State int
type StateChange int
type StateChangeReturn int

// Maps to GstState enumeration
const (
	StateVoidPending State = 0
	StateNull        State = 1
	StateReady       State = 2
	StatePaused      State = 3
	StatePlaying     State = 4
)

// Maps to GstStateChange enumeration
const (
	StateChangeNullToReady      StateChange = 10
	StateChangeReadyToPaused    StateChange = 19
	StateChangePausedToPlaying  StateChange = 28
	StateChangePlayingToPaused  StateChange = 35
	StateChangePausedToReady    StateChange = 26
	StateChangeReadyToNull      StateChange = 17
	StateChangeNullToNull       StateChange = 9
	StateChangeReadyToReady     StateChange = 18
	StateChangePausedToPaused   StateChange = 27
	StateChangePlayingToPlaying StateChange = 36
)

// Maps to GstStateChangeReturn enumeration
const (
	StateChangeReturnFailure   StateChangeReturn = 0
	StateChangeReturnSuccess   StateChangeReturn = 1
	StateChangeReturnAsync     StateChangeReturn = 2
	StateChangeReturnNoPreroll StateChangeReturn = 3
)
