package updater

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/endpoint"
)

// ManifestURL is the signed update manifest attached to the latest stable
// release. GitHub's /latest/ alias skips drafts and pre-releases.
const ManifestURL = "https://github.com/AnoRebel/aDex-UI/releases/latest/download/update-manifest.json"

const releasePageBase = "https://github.com/AnoRebel/aDex-UI/releases/tag/v"

// Info is the updater's status as the frontend sees it.
type Info struct {
	// Available is false when the updater could not be configured; Reason
	// then says why. The app runs normally either way.
	Available      bool   `json:"available"`
	Reason         string `json:"reason,omitempty"`
	CurrentVersion string `json:"currentVersion"`
	// CanSelfUpdate is false for installs that can only be pointed at the
	// release page; SelfUpdateReason says why.
	CanSelfUpdate    bool   `json:"canSelfUpdate"`
	SelfUpdateReason string `json:"selfUpdateReason,omitempty"`
	State            string `json:"state"`
}

// Release describes a newer version found by Check.
type Release struct {
	Version     string    `json:"version"`
	Name        string    `json:"name"`
	Notes       string    `json:"notes"`
	PublishedAt time.Time `json:"publishedAt"`
	// URL is the release page, for installs that can't update in-app.
	URL string `json:"url"`
}

// Service exposes the Wails updater to the frontend. It runs headless: the
// frontend drives each step and follows progress through the
// wails:updater:* events.
type Service struct {
	u           *wupdater.Updater // nil when unavailable
	version     string
	unavailable string

	// Eligibility decides whether Install may start. Tests replace it.
	Eligibility func() (bool, string)
}

// NewService configures u to read the signed manifest at manifestURL and
// accept only artifacts signed by pubKey. It always returns a usable service:
// on failure the service reports itself unavailable, with the reason, and the
// error is returned for logging. A broken updater never stops the app.
//
// This is a constructor rather than a method so it isn't bound to the
// frontend.
func NewService(u *wupdater.Updater, version string, pubKey []byte, manifestURL string) (*Service, error) {
	s := &Service{version: version, Eligibility: CanSelfUpdate}
	err := s.init(u, version, pubKey, manifestURL)
	if err != nil {
		s.unavailable = err.Error()
	} else {
		s.u = u
	}
	return s, err
}

func (s *Service) init(u *wupdater.Updater, version string, pubKey []byte, manifestURL string) error {
	switch {
	case u == nil:
		return errors.New("no updater")
	case version == "":
		return errors.New("unknown app version")
	case len(pubKey) == 0:
		return errors.New("no update signing key")
	}
	p, err := endpoint.New(endpoint.Config{URL: manifestURL, HTTPClient: httpClient()})
	if err != nil {
		return err
	}
	return u.Init(wupdater.Config{
		CurrentVersion: version,
		Providers:      []wupdater.Provider{p},
		PublicKey:      pubKey,
		Window:         wupdater.WindowNone,
	})
}

// httpClient bounds connecting and waiting for headers, not the whole
// request: the endpoint provider's default 30-second total timeout would cut
// off a large download on a slow link. A 404 is turned into an error so a
// release without a manifest reads as a failed check, not "up to date".
func httpClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = (&net.Dialer{Timeout: 30 * time.Second}).DialContext
	tr.ResponseHeaderTimeout = 30 * time.Second
	return &http.Client{Transport: notFoundIsError{tr}}
}

type notFoundIsError struct{ base http.RoundTripper }

func (t notFoundIsError) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(r)
	if err == nil && resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: not found", r.URL)
	}
	return resp, err
}

func (s *Service) updater() (*wupdater.Updater, error) {
	if s.u == nil {
		return nil, fmt.Errorf("updates unavailable: %s", s.unavailable)
	}
	return s.u, nil
}

// Info reports the running version, whether updates work at all, and
// whether this install can replace itself.
func (s *Service) Info() Info {
	info := Info{Available: s.u != nil, Reason: s.unavailable, CurrentVersion: s.version, State: string(wupdater.StateUnconfigured)}
	if s.u != nil {
		info.State = string(s.u.State())
	}
	info.CanSelfUpdate, info.SelfUpdateReason = s.Eligibility()
	return info
}

// Check looks for a newer stable release. It returns nil when up to date.
func (s *Service) Check(ctx context.Context) (*Release, error) {
	u, err := s.updater()
	if err != nil {
		return nil, err
	}
	rel, err := u.Check(ctx)
	if err != nil || rel == nil {
		return nil, err
	}
	return &Release{
		Version:     rel.Version,
		Name:        rel.Name,
		Notes:       rel.Notes,
		PublishedAt: rel.PublishedAt,
		URL:         releasePageBase + rel.Version,
	}, nil
}

// Install downloads, verifies and stages the release found by the last
// Check. It refuses before downloading anything when this install can't
// replace itself.
func (s *Service) Install(ctx context.Context) error {
	u, err := s.updater()
	if err != nil {
		return err
	}
	if ok, reason := s.Eligibility(); !ok {
		return fmt.Errorf("self-update unavailable: %s", reason)
	}
	return u.DownloadAndInstall(ctx)
}

// Restart quits and relaunches into the staged update.
func (s *Service) Restart(ctx context.Context) error {
	u, err := s.updater()
	if err != nil {
		return err
	}
	return u.Restart(ctx)
}
