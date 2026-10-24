package strategy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestCheckPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "plain keys", path: "mcpServers.notes"},
		{name: "escaped dot", path: `projects./Users/test/my\.repo`},
		{name: "empty", path: "", wantErr: "owned path is empty"},
		{name: "wildcard", path: "mcpServers.*", wantErr: "uses query syntax"},
		{name: "array query", path: "servers.#", wantErr: "uses query syntax"},
		{name: "modifier", path: "mcpServers|@reverse", wantErr: "uses query syntax"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := checkPath(tc.path)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestHashIgnoresFormatting(t *testing.T) {
	t.Parallel()

	compactValue := gjson.GetBytes(fixture(t, "applied-set.json"), notes)
	indentedValue := gjson.GetBytes(fixture(t, "source.json"), notes)

	assert.Equal(t, hashValue(indentedValue), hashValue(compactValue))
	assert.NotEqual(t, hashValue(indentedValue), hashValue(gjson.GetBytes(fixture(t, "source-changed.json"), notes)))
	assert.Empty(t, hashValue(gjson.GetBytes(fixture(t, "target-without.json"), notes)), "an absent path has no hash")
}
