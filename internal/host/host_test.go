package host

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	t.Parallel()

	lookup := func() (string, error) { return "MacBook-Pro\n", nil }
	failing := func() (string, error) { return "", errors.New("scutil failed") }

	tests := []struct {
		name    string
		flag    string
		file    *string
		lookup  func() (string, error)
		want    string
		wantErr string
	}{
		{name: "flag wins over file", flag: "cli", file: new("work\n"), lookup: failing, want: "cli"},
		{name: "file wins over lookup", file: new("work\n"), lookup: failing, want: "work"},
		{name: "lookup without file", lookup: lookup, want: "MacBook-Pro"},
		{name: "empty file", file: new(" \n"), lookup: lookup, wantErr: "is empty"},
		{name: "lookup error", lookup: failing, wantErr: "scutil failed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := filepath.Join(t.TempDir(), "host")
			if tc.file != nil {
				require.NoError(t, os.WriteFile(file, []byte(*tc.file), 0o644))
			}

			got, err := Name(tc.flag, file, tc.lookup)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
