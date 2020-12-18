package riff

import (
	"sync"
)

// SampleLoop represents sample loop used in ChunkSMPL.
//
// Source:
// https://sites.google.com/site/musicgapi/technical-documents/wav-file-format
type SampleLoop struct {
	// The Cue Point ID specifies the unique ID that corresponds to one
	// of the defined cue points in the cue point list. Furthermore,
	// this ID corresponds to any labels defined in the associated data
	// list chunk which allows text labels to be assigned to the various
	// sample loops.
	CuePointID uint32

	// The type field defines how the waveform samples will be looped.
	// Loop types:
	//
	// - 0 - Loop forward (normal)
	// - 1 - Alternating loop (forward/backward, also known as Ping Pong)
	// - 2 - Loop backward (reverse)
	// - 3 - 31	Reserved for future standard types
	// - 32 - 0xFFFFFFFF Sampler specific types (defined by manufacturer)
	//
	Type uint32

	// The start value specifies the byte offset into the waveform data
	// of the first sample to be played in the loop.
	Start uint32

	// The end value specifies the byte offset into the waveform data of
	// the last sample to be played in the loop.
	End uint32

	// The fractional value specifies a fraction of a sample at which to loop.
	// This allows a loop to be fine tuned at a resolution greater than one
	// sample. The value can range from 0x00000000 to 0xFFFFFFFF.
	// A value of 0 means no fraction,
	// a value of 0x80000000 means 1/2 of a sample length.
	// The 0xFFFFFFFF is the smallest fraction of a sample that
	// can be represented.
	Fraction uint32

	// The play count value determines the number of times to play the loop.
	// A value of 0 specifies an infinite sustain loop. An infinite sustain
	// loop will continue looping until some external force interrupts
	// playback, such as the musician releasing the key that triggered the
	// wave's playback. All other values specify an absolute number of
	// times to loop.
	PlayCnt uint32
}

func (sl *SampleLoop) Reset() {
	sl.CuePointID = 0
	sl.Type = 0
	sl.Start = 0
	sl.End = 0
	sl.Fraction = 0
	sl.PlayCnt = 0
}

// sampleLoopPool is a pool for SampleLoop instances.
var sampleLoopPool = &sync.Pool{
	New: func() interface{} {
		return &SampleLoop{}
	},
}
