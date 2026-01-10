package engine

// Result contains the output of a successful transcription.
type Result struct {
	// Transcript is the full text transcription.
	Transcript string

	// Segments are the timestamped text segments.
	Segments []Segment
}

// Segment represents a timestamped portion of the transcript.
type Segment struct {
	// Start is the start time in seconds.
	Start float64

	// End is the end time in seconds.
	End float64

	// Text is the transcribed text for this segment.
	Text string
}
