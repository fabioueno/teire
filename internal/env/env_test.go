package env

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		stateHome  string
		configHome string
		want       Env
	}{
		{
			name: "XDG defaults",
			want: Env{
				Home:      "/Users/test",
				StateDir:  "/Users/test/.local/state/teire",
				ConfigDir: "/Users/test/.config/teire",
			},
		},
		{
			name:       "XDG variables set",
			stateHome:  "/var/state",
			configHome: "/etc/xdg",
			want: Env{
				Home:      "/Users/test",
				StateDir:  "/var/state/teire",
				ConfigDir: "/etc/xdg/teire",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, New("/Users/test", tc.stateHome, tc.configHome))
		})
	}
}

func TestPaths(t *testing.T) {
	t.Parallel()

	e := New("/Users/test", "", "")
	now := time.Date(2026, 10, 24, 12, 30, 5, 0, time.FixedZone("CEST", 2*60*60))

	assert.Equal(t, "/Users/test/.local/state/teire/state.json", e.StatePath())
	assert.Equal(t, "/Users/test/.config/teire/host", e.HostFile())
	assert.Equal(t, "/Users/test/.local/state/teire/backups/20261024T103005Z", e.Backups(now))
}
