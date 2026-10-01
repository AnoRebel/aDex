package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

// Every shipped cue must decode. The cues span several WAV encodings
// (24000/44100/48000 Hz, 16/24/32-bit, integer and float), so this is the
// check that the decoder handles the real asset set rather than one sample.
func TestDecodeCue_AllCuesDecode(t *testing.T) {
	for id := range CueNames {
		pcm, err := decodeCue(id)
		if err != nil {
			t.Errorf("cue %q: %v", id, err)
			continue
		}
		if len(pcm) == 0 {
			t.Errorf("cue %q decoded to zero bytes", id)
			continue
		}
		// float32 stereo => 8 bytes per frame.
		if len(pcm)%(4*cueChannels) != 0 {
			t.Errorf("cue %q: %d bytes is not a whole number of stereo float32 frames", id, len(pcm))
		}
	}
}

func TestDecodeCue_UnknownCue(t *testing.T) {
	if _, err := decodeCue("does-not-exist"); err == nil {
		t.Fatal("expected an error for an unknown cue id")
	}
}

// The second decode must come from cache and be identical.
func TestDecodeCue_IsCached(t *testing.T) {
	a, err := decodeCue("scan")
	if err != nil {
		t.Fatalf("first decode: %v", err)
	}
	b, err := decodeCue("scan")
	if err != nil {
		t.Fatalf("second decode: %v", err)
	}
	if len(a) != len(b) {
		t.Fatalf("cached decode differs: %d vs %d bytes", len(a), len(b))
	}
}

// Decoded audio must land in -1..1. IEEE-float WAVs (format 3) previously
// decoded to raw float BIT PATTERNS widened into ints — values around +/-2e9 —
// which played as pure noise. Integer-PCM cues were unaffected, which is why
// only the click and keypress sounds were correct.
func TestDecodeCue_SamplesAreNormalised(t *testing.T) {
	for id := range CueNames {
		pcm, err := decodeCue(id)
		if err != nil {
			t.Errorf("cue %q: %v", id, err)
			continue
		}
		var peak float32
		for i := 0; i+4 <= len(pcm); i += 4 {
			v := math.Float32frombits(binary.LittleEndian.Uint32(pcm[i : i+4]))
			if v < 0 {
				v = -v
			}
			if v > peak {
				peak = v
			}
		}
		if peak > 1.0001 {
			t.Errorf("cue %q: peak amplitude %g exceeds full scale — samples are not normalised", id, peak)
		}
		if peak == 0 {
			t.Errorf("cue %q: decoded to silence", id)
		}
	}
}
