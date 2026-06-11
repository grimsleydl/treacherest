package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportCoupRoleImagesCopiesSupportedFilesToCanonicalIDs(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := t.TempDir()

	sourceFiles := map[string][]byte{
		"king.jpg":             {0xff, 0xd8, 0xff, 0xd9},
		"blue-knight.png":      {'p', 'n', 'g'},
		"black-knight.webp":    {'w', 'e', 'b', 'p'},
		"red-knight.jpeg":      {0xff, 0xd8, 'r', 0xff, 0xd9},
		"green-knight.jpg":     {0xff, 0xd8, 'g', 0xff, 0xd9},
		"wasteland-knight.jpg": {0xff, 0xd8, 'w', 0xff, 0xd9},
	}
	for name, data := range sourceFiles {
		if err := os.WriteFile(filepath.Join(sourceDir, name), data, 0644); err != nil {
			t.Fatalf("write source %s: %v", name, err)
		}
	}

	result, err := ImportCoupRoleImages(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("import Coup role images: %v", err)
	}

	if result.Imported != 6 {
		t.Fatalf("expected 6 imports, got %d", result.Imported)
	}
	assertFileBytes(t, filepath.Join(outputDir, "1001.jpg"), sourceFiles["king.jpg"])
	assertFileBytes(t, filepath.Join(outputDir, "1002.png"), sourceFiles["blue-knight.png"])
	assertFileBytes(t, filepath.Join(outputDir, "1003.webp"), sourceFiles["black-knight.webp"])
	assertFileBytes(t, filepath.Join(outputDir, "1004.jpeg"), sourceFiles["red-knight.jpeg"])
	assertFileBytes(t, filepath.Join(outputDir, "1005.jpg"), sourceFiles["green-knight.jpg"])
	assertFileBytes(t, filepath.Join(outputDir, "1006.jpg"), sourceFiles["wasteland-knight.jpg"])
}

func TestImportCoupRoleImagesRemovesStaleCanonicalFilesForRole(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := t.TempDir()

	for _, name := range []string{"king.png", "blue-knight.jpg", "black-knight.jpg", "red-knight.jpg", "green-knight.jpg", "wasteland-knight.jpg"} {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(name), 0644); err != nil {
			t.Fatalf("write source %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(outputDir, "1001.jpg"), []byte("stale"), 0644); err != nil {
		t.Fatalf("write stale output: %v", err)
	}

	if _, err := ImportCoupRoleImages(sourceDir, outputDir); err != nil {
		t.Fatalf("import Coup role images: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "1001.jpg")); !os.IsNotExist(err) {
		t.Fatalf("expected stale 1001.jpg to be removed, got err %v", err)
	}
	assertFileBytes(t, filepath.Join(outputDir, "1001.png"), []byte("king.png"))
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != string(want) {
		t.Fatalf("expected %s bytes %q, got %q", path, string(want), string(got))
	}
}
