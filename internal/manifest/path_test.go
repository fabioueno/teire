package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		target  string
		want    string
		wantErr string
	}{
		{name: "home itself", target: "~", want: home},
		{name: "below home", target: "~/.config/zed", want: home + "/.config/zed"},
		{name: "absolute", target: "/etc/xdg", want: "/etc/xdg"},
		{name: "cleaned", target: "~/.config//zed/", want: home + "/.config/zed"},
		{name: "relative", target: ".config", wantErr: `target ".config" is not absolute`},
		{name: "another user's home", target: "~root/.config", wantErr: "is not absolute"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := expandTarget(tc.target, home)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
