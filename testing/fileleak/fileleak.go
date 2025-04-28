//go:build !solution

package fileleak

import (
	"os"
	"path/filepath"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

func getOpenFDs() map[string]bool {
	result := make(map[string]bool)

	path := "/proc/self/fd"

	entries, _ := os.ReadDir(path)

	for _, entry := range entries {
		if !entry.IsDir() {
			description, _ := os.Readlink(filepath.Join(path, entry.Name()))
			result[entry.Name()+description] = true
		}
	}
	return result
}

func VerifyNone(t testingT) {
	initialFDs := getOpenFDs()

	t.Cleanup(func() {
		currentFDs := getOpenFDs()

		for fd := range currentFDs {
			if !initialFDs[fd] {
				t.Errorf("oh no")
			}
		}
	})
}
