package cli

import (
	"bytes"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionPrintsTheBuildVersion(t *testing.T) {
	root := NewRoot()
	var out bytes.Buffer

	root.SetOut(&out)
	root.SetArgs([]string{"version"})

	require.NoError(t, root.Execute())
	assert.Equal(t, "dev\n", out.String())
}

func TestVersionRejectsArguments(t *testing.T) {
	root := NewRoot()
	root.SetArgs([]string{"version", "extra"})

	assert.Error(t, root.Execute())
}

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	module := func(v string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Path: "github.com/fabioueno/teire", Version: v}}
	}

	tests := []struct {
		name    string
		stamped string
		info    *debug.BuildInfo
		ok      bool
		want    string
	}{
		{name: "stamped by the formula", stamped: "0.1.0", info: module("v0.1.0"), ok: true, want: "0.1.0"},
		{name: "go install of a tag", stamped: "dev", info: module("v0.2.0"), ok: true, want: "0.2.0"},
		{name: "go install of a commit", stamped: "dev", info: module("v0.2.1-0.20261024120000-abcdef123456"), ok: true, want: "0.2.1-0.20261024120000-abcdef123456"},
		{name: "build from a checkout", stamped: "dev", info: module("(devel)"), ok: true, want: "dev"},
		{name: "no build info", stamped: "dev", ok: false, want: "dev"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, resolveVersion(tc.stamped, tc.info, tc.ok))
		})
	}
}
