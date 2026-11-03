package tests

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"aDex/internal/services/updater"
	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
)

func readOnlyDir(t *testing.T) string {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	return dir
}

func TestEligibility(t *testing.T) {
	rw := t.TempDir()
	ro := readOnlyDir(t)

	cases := []struct {
		name, goos, exe, appImage string
		want                      bool
		reason                    string // substring of the reason when not eligible
	}{
		{"linux appimage, writable", "linux", "/tmp/.mount_x/usr/bin/adex", filepath.Join(rw, "aDex.AppImage"), true, ""},
		{"linux no appimage (distro or raw binary)", "linux", filepath.Join(rw, "adex"), "", false, "AppImage"},
		{"linux appimage, read-only", "linux", "/tmp/.mount_x/usr/bin/adex", filepath.Join(ro, "aDex.AppImage"), false, "not writable"},
		{"darwin bundle, writable parent", "darwin", filepath.Join(rw, "aDex.app/Contents/MacOS/adex"), "", true, ""},
		{"darwin bundle, read-only parent", "darwin", filepath.Join(ro, "aDex.app/Contents/MacOS/adex"), "", false, "not writable"},
		{"darwin binary outside a bundle", "darwin", filepath.Join(rw, "adex"), "", false, ".app bundle"},
		{"windows writable dir", "windows", filepath.Join(rw, "adex.exe"), "", true, ""},
		{"windows read-only dir", "windows", filepath.Join(ro, "adex.exe"), "", false, "not writable"},
		{"unsupported os", "freebsd", filepath.Join(rw, "adex"), "", false, "not supported"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ok, reason := updater.Eligibility(c.goos, c.exe, c.appImage)
			if ok != c.want {
				t.Fatalf("eligible = %v (%q), want %v", ok, reason, c.want)
			}
			if !ok && !strings.Contains(reason, c.reason) {
				t.Fatalf("reason %q does not mention %q", reason, c.reason)
			}
			if ok && reason != "" {
				t.Fatalf("eligible but reason %q", reason)
			}
		})
	}

	// The probe must not leave files behind.
	if entries, _ := os.ReadDir(rw); len(entries) != 0 {
		t.Fatalf("probe left %d entries in %s", len(entries), rw)
	}
}

// --- updater service --------------------------------------------------------

// stubHost satisfies wupdater.Host without a running Wails app.
type stubHost struct {
	mu     sync.Mutex
	events []string
}

func (h *stubHost) Emit(name string, _ ...any) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, name)
	return true
}
func (h *stubHost) OnEvent(string, func(any)) func()                        { return func() {} }
func (h *stubHost) OpenWindow(wupdater.WindowOptions) wupdater.WindowHandle { return nil }
func (h *stubHost) Quit()                                                   {}

// releaseServer serves a manifest for version with one artifact for the
// running platform. sign controls whether the entry carries a signature;
// serve is the bytes actually returned for the artifact (which may differ
// from what was signed).
func releaseServer(t *testing.T, priv ed25519.PrivateKey, version string, signed, serve []byte, sign bool) *httptest.Server {
	t.Helper()
	sum := sha512.Sum512(signed)
	art := map[string]any{
		"url": "/artifact", "platform": runtime.GOOS, "arch": runtime.GOARCH,
		"digestAlgo": "sha512", "digest": base64.StdEncoding.EncodeToString(sum[:]),
	}
	if sign {
		sig, err := priv.Sign(nil, sum[:], &ed25519.Options{Hash: crypto.SHA512})
		if err != nil {
			t.Fatal(err)
		}
		art["signatureAlgo"] = "ed25519ph"
		art["signature"] = base64.StdEncoding.EncodeToString(sig)
	}
	manifest, _ := json.Marshal(map[string]any{
		"schemaVersion": 1, "version": version, "name": "aDex " + version,
		"notes": "notes", "artifacts": []any{art},
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/update-manifest.json", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(manifest) })
	mux.HandleFunc("/artifact", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(serve) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestKey(t *testing.T) (ed25519.PrivateKey, []byte) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKIXPublicKey(pub)
	return priv, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func newService(t *testing.T, version string, pub []byte, url string, eligible bool) *updater.Service {
	t.Helper()
	s, err := updater.NewService(wupdater.New(&stubHost{}), version, pub, url)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	s.Eligibility = func() (bool, string) { return eligible, "test" }
	return s
}

func TestUpdaterInitFailureLeavesServiceUnavailable(t *testing.T) {
	_, pub := newTestKey(t)
	s, err := updater.NewService(wupdater.New(&stubHost{}), "", pub, "http://127.0.0.1/m.json")
	if err == nil {
		t.Fatal("NewService with an empty version should fail")
	}
	info := s.Info()
	if info.Available || info.Reason == "" {
		t.Fatalf("Info = %+v, want unavailable with a reason", info)
	}
	if _, err := s.Check(context.Background()); err == nil {
		t.Fatal("Check on an unavailable updater should fail")
	}
}

func TestUpdaterCheck(t *testing.T) {
	priv, pub := newTestKey(t)
	payload := []byte("new build")

	t.Run("newer version is reported", func(t *testing.T) {
		srv := releaseServer(t, priv, "2.0.0", payload, payload, true)
		s := newService(t, "1.0.0", pub, srv.URL+"/update-manifest.json", true)
		rel, err := s.Check(context.Background())
		if err != nil || rel == nil {
			t.Fatalf("Check = (%v, %v), want a release", rel, err)
		}
		if rel.Version != "2.0.0" || rel.Name != "aDex 2.0.0" || !strings.HasSuffix(rel.URL, "/releases/tag/v2.0.0") {
			t.Fatalf("release = %+v", rel)
		}
		if got := s.Info().State; got != "available" {
			t.Fatalf("state = %q, want available", got)
		}
	})

	t.Run("equal version is up to date", func(t *testing.T) {
		srv := releaseServer(t, priv, "1.0.0", payload, payload, true)
		s := newService(t, "1.0.0", pub, srv.URL+"/update-manifest.json", true)
		rel, err := s.Check(context.Background())
		if err != nil || rel != nil {
			t.Fatalf("Check = (%v, %v), want up to date", rel, err)
		}
	})

	t.Run("missing manifest is an error", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		t.Cleanup(srv.Close)
		s := newService(t, "1.0.0", pub, srv.URL+"/update-manifest.json", true)
		if rel, err := s.Check(context.Background()); err == nil {
			t.Fatalf("Check = (%v, nil), want an error for a 404 manifest", rel)
		}
	})
}

func TestUpdaterInstallVerification(t *testing.T) {
	priv, pub := newTestKey(t)
	payload := []byte("new build")

	install := func(t *testing.T, srv *httptest.Server) error {
		s := newService(t, "1.0.0", pub, srv.URL+"/update-manifest.json", true)
		if rel, err := s.Check(context.Background()); err != nil || rel == nil {
			t.Fatalf("Check = (%v, %v)", rel, err)
		}
		return s.Install(context.Background())
	}

	t.Run("unsigned entry is refused", func(t *testing.T) {
		err := install(t, releaseServer(t, priv, "2.0.0", payload, payload, false))
		if err == nil || !strings.Contains(err.Error(), "signature") {
			t.Fatalf("Install = %v, want a signature error", err)
		}
	})

	t.Run("tampered artifact is refused", func(t *testing.T) {
		err := install(t, releaseServer(t, priv, "2.0.0", payload, []byte("evil build"), true))
		if err == nil || !strings.Contains(err.Error(), "digest mismatch") {
			t.Fatalf("Install = %v, want a digest mismatch", err)
		}
	})

	t.Run("signature from another key is refused", func(t *testing.T) {
		other, _ := newTestKey(t)
		err := install(t, releaseServer(t, other, "2.0.0", payload, payload, true))
		if err == nil || !strings.Contains(err.Error(), "did not verify") {
			t.Fatalf("Install = %v, want a signature error", err)
		}
	})
}

func TestUpdaterInstallRefusedWhenIneligible(t *testing.T) {
	priv, pub := newTestKey(t)
	payload := []byte("new build")
	var downloads int
	srv := releaseServer(t, priv, "2.0.0", payload, payload, true)
	counting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/artifact" {
			downloads++
		}
		resp, err := http.Get(srv.URL + r.URL.Path)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()
		_, _ = io.Copy(w, resp.Body)
	}))
	t.Cleanup(counting.Close)

	s := newService(t, "1.0.0", pub, counting.URL+"/update-manifest.json", false)
	if rel, err := s.Check(context.Background()); err != nil || rel == nil {
		t.Fatalf("Check = (%v, %v)", rel, err)
	}
	if err := s.Install(context.Background()); err == nil || !strings.Contains(err.Error(), "self-update unavailable") {
		t.Fatalf("Install = %v, want refusal", err)
	}
	if downloads != 0 {
		t.Fatalf("artifact downloaded %d times, want 0", downloads)
	}
}
