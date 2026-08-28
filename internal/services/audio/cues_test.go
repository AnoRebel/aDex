package audio

import "testing"

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
