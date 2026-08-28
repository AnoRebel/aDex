package audio

import (
	"bytes"
	"embed"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-audio/wav"
)

// Cue audio is played by Go rather than the webview.
//
// The webview refuses to start audio before a user gesture (WebKitGTK's
// media-playback-requires-user-gesture, which Wails exposes no setting for on
// Linux), so splash cues scheduled during boot were silently dropped — the
// boot sequence played nothing until the user first clicked or typed. Playing
// through the OS audio stack from Go sidesteps that policy entirely: there is
// no webview involved and therefore nothing to unlock.
//
//go:embed cues/*.wav
var cueFS embed.FS

// Context format. Decoded audio is converted to match, because the shipped
// cues are NOT uniform: they span 24000/44100/48000 Hz, 16/24/32-bit, and both
// integer PCM and float encodings. Feeding a raw WAV to oto (which expects
// bare PCM in exactly this format) would play the file header as noise.
const (
	cueSampleRate = 44100
	cueChannels   = 2
)

var (
	cueCacheMu sync.RWMutex
	cueCache   = map[string][]byte{}
)

// CueNames maps a cue id to its embedded file. Ids match the frontend's cue
// vocabulary so the two stay aligned.
var CueNames = map[string]string{
	"alarm":       "alarm.wav",
	"denied":      "denied.wav",
	"error":       "error.wav",
	"expand":      "expand.wav",
	"folder":      "folder.wav",
	"granted":     "granted.wav",
	"info":        "info.wav",
	"keyboard":    "keyboard.wav",
	"panels":      "panels.wav",
	"scan":        "scan.wav",
	"stdin":       "stdin.wav",
	"stdout":      "stdout.wav",
	"theme":       "theme.wav",
	"keypress":    "mechanical-keyboard.wav",
	"destructive": "handgun-click.wav",
	"gunshot":     "gunshot.wav",
	"click":       "click-tone.wav",
}

// decodeCue returns the cue as float32 little-endian PCM at the context's
// sample rate and channel count, ready to hand straight to oto. Results are
// cached: decoding is pure CPU and a cue may play many times per session.
func decodeCue(id string) ([]byte, error) {
	cueCacheMu.RLock()
	if pcm, ok := cueCache[id]; ok {
		cueCacheMu.RUnlock()
		return pcm, nil
	}
	cueCacheMu.RUnlock()

	file, ok := CueNames[id]
	if !ok {
		return nil, fmt.Errorf("unknown cue %q", id)
	}
	raw, err := cueFS.ReadFile("cues/" + file)
	if err != nil {
		return nil, fmt.Errorf("cue %q: %w", id, err)
	}

	dec := wav.NewDecoder(bytes.NewReader(raw))
	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("decode cue %q: %w", id, err)
	}

	// Normalise integer PCM to -1..1. Float-encoded WAVs already are.
	scale := 1.0
	if buf.SourceBitDepth > 0 && dec.WavAudioFormat != 3 {
		scale = 1.0 / float64(int64(1)<<(buf.SourceBitDepth-1))
	}

	srcCh := buf.Format.NumChannels
	if srcCh < 1 {
		srcCh = 1
	}
	frames := len(buf.Data) / srcCh

	// Linear resample when the cue's rate differs from the context's. These
	// are short interface sounds, so nearest-neighbour is inaudible here and
	// avoids pulling in a resampling dependency.
	ratio := float64(buf.Format.SampleRate) / float64(cueSampleRate)
	outFrames := int(float64(frames) / ratio)

	out := make([]byte, 0, outFrames*cueChannels*4)
	var scratch [4]byte
	for i := 0; i < outFrames; i++ {
		src := int(float64(i) * ratio)
		if src >= frames {
			break
		}
		for c := 0; c < cueChannels; c++ {
			// Mono sources feed both output channels.
			sc := c
			if sc >= srcCh {
				sc = srcCh - 1
			}
			v := float64(buf.Data[src*srcCh+sc]) * scale
			if v > 1 {
				v = 1
			} else if v < -1 {
				v = -1
			}
			binary.LittleEndian.PutUint32(scratch[:], math.Float32bits(float32(v)))
			out = append(out, scratch[:]...)
		}
	}

	cueCacheMu.Lock()
	cueCache[id] = out
	cueCacheMu.Unlock()
	return out, nil
}

// PlayCue plays a named cue at the given volume (0..1).
//
// It returns nil when audio is unavailable rather than an error: cues are
// decorative, and a missing sound device must never surface as a failure in
// the caller's path.
func (s *Service) PlayCue(id string, volume float64) error {
	if !s.isInitialized || s.otoCtx == nil {
		return nil
	}
	if volume <= 0 {
		return nil
	}
	if volume > 1 {
		volume = 1
	}

	pcm, err := decodeCue(id)
	if err != nil {
		return err
	}

	player := s.otoCtx.NewPlayer(bytes.NewReader(pcm))
	if player == nil {
		return nil
	}
	player.SetVolume(volume)

	playerID := fmt.Sprintf("cue-%s-%p", id, player)
	s.mu.Lock()
	s.players[playerID] = player
	s.mu.Unlock()

	player.Play()

	// Reap when finished. Close under the lock and only if this is still the
	// registered player, so a concurrent StopAllSounds/Shutdown cannot close
	// the same C-backed player twice.
	go func() {
		for player.IsPlaying() {
			time.Sleep(20 * time.Millisecond)
		}
		s.mu.Lock()
		if cur, ok := s.players[playerID]; ok && cur == player {
			player.Close()
			delete(s.players, playerID)
		}
		s.mu.Unlock()
	}()

	return nil
}

// AvailableCues lists the cue ids this service can play.
func (s *Service) AvailableCues() []string {
	ids := make([]string, 0, len(CueNames))
	for id := range CueNames {
		ids = append(ids, id)
	}
	return ids
}
