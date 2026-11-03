package tests

import (
	"os"
	"regexp"
	"testing"

	"aDex/internal/services/updater"
)

// The backend's version must come from build/config.yml (spec: build-and-dev,
// "Project metadata has a single source of truth").
func TestParseVersionFromRealConfig(t *testing.T) {
	data, err := os.ReadFile("../../build/config.yml")
	if err != nil {
		t.Fatal(err)
	}
	v, err := updater.ParseVersion(data)
	if err != nil {
		t.Fatalf("real config: %v", err)
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+([-+].+)?$`).MatchString(v) {
		t.Fatalf("version %q is not bare semver", v)
	}
	// Cross-check with a plain text match, independent of the YAML parser.
	m := regexp.MustCompile(`(?m)^info:\n(?:[ \t]+.*\n)*?[ \t]+version:[ \t]*"?v?([^"\n]+)"?`).FindSubmatch(data)
	if m == nil || string(m[1]) != v {
		t.Fatalf("ParseVersion = %q, info.version in file = %q", v, m)
	}
}

func TestParseVersion(t *testing.T) {
	cases := []struct {
		name, yaml, want string
		wantErr          bool
	}{
		{"plain", "info:\n  version: \"2.3.4\"\n", "2.3.4", false},
		{"v prefix stripped", "info:\n  version: v2.3.4\n", "2.3.4", false},
		{"empty version", "info:\n  version: \"\"\n", "", true},
		{"missing info", "version: '3'\n", "", true},
		{"malformed yaml", "info: [unclosed\n", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := updater.ParseVersion([]byte(c.yaml))
			if (err != nil) != c.wantErr || got != c.want {
				t.Fatalf("got (%q, %v), want (%q, err=%v)", got, err, c.want, c.wantErr)
			}
		})
	}
}
