package testutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestGetTestDBPathAndCleanup(t *testing.T) {
	const pathCount = 32
	paths := make([]string, pathCount)
	var wg sync.WaitGroup

	for i := range paths {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			paths[i] = GetTestDBPath()
		}(i)
	}
	wg.Wait()
	defer func() {
		for _, path := range paths {
			_ = os.Remove(path)
		}
	}()

	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if filepath.Dir(path) != os.TempDir() {
			t.Errorf("GetTestDBPath() = %q, want path under %q", path, os.TempDir())
		}
		if _, exists := seen[path]; exists {
			t.Errorf("GetTestDBPath() returned duplicate path %q", path)
		}
		seen[path] = struct{}{}
		if err := os.WriteFile(path, nil, 0600); err != nil {
			t.Fatalf("create test database %q: %v", path, err)
		}
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	isolatedDir := t.TempDir()
	if err := os.Chdir(isolatedDir); err != nil {
		t.Fatalf("change to isolated working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	}()

	if err := CleanupTestDB(); err != nil {
		t.Fatalf("CleanupTestDB(): %v", err)
	}
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("database %q still exists or stat failed: %v", path, err)
		}
	}
}
