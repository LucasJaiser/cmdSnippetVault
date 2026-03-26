package cli

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func goldenFile(t *testing.T, name string, actual string) string {
	t.Helper()

	path := filepath.Join("testdata", name+".golden")

	if *update {
		err := os.WriteFile(path, []byte(actual), 0o644)
		require.NoError(t, err, "failed to update golden file")
		return actual
	}

	expected, err := os.ReadFile(path)
	require.NoError(t, err, "golden file not found, run with -update to create")
	return string(expected)
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	require.NoError(t, err)
	r.Close()

	return string(out)
}
