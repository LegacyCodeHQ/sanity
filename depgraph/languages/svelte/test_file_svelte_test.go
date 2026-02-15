package svelte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{
			name:     "test suffix",
			filePath: "/project/src/App.test.svelte",
			want:     true,
		},
		{
			name:     "spec suffix",
			filePath: "/project/src/Button.spec.svelte",
			want:     true,
		},
		{
			name:     "__tests__ directory",
			filePath: "/project/src/__tests__/Header.svelte",
			want:     true,
		},
		{
			name:     "regular svelte file",
			filePath: "/project/src/App.svelte",
			want:     false,
		},
		{
			name:     "non-svelte file",
			filePath: "/project/src/App.test.js",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsTestFile(tc.filePath))
		})
	}
}
