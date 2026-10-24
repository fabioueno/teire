package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantShow bool
	}{
		{name: "plain error", err: errors.New("boom"), wantCode: 1, wantShow: true},
		{name: "silent exit", err: &exitError{code: 2}, wantCode: 2, wantShow: false},
		{name: "exit with a message", err: &exitError{code: 2, err: errors.New("1 conflict left")}, wantCode: 2, wantShow: true},
		{name: "wrapped exit", err: fmt.Errorf("apply: %w", &exitError{code: 1}), wantCode: 1, wantShow: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code, show := ExitCode(tc.err)

			assert.Equal(t, tc.wantCode, code)
			assert.Equal(t, tc.wantShow, show)
		})
	}
}
