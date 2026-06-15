package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportCoupRoleImagesCopiesSupportedFilesToCanonicalSlugs(t *testing.T) {
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
	assertFileBytes(t, filepath.Join(outputDir, "king.jpg"), sourceFiles["king.jpg"])
	assertFileBytes(t, filepath.Join(outputDir, "blue-knight.png"), sourceFiles["blue-knight.png"])
	assertFileBytes(t, filepath.Join(outputDir, "black-knight.webp"), sourceFiles["black-knight.webp"])
	assertFileBytes(t, filepath.Join(outputDir, "red-knight.jpeg"), sourceFiles["red-knight.jpeg"])
	assertFileBytes(t, filepath.Join(outputDir, "green-knight.jpg"), sourceFiles["green-knight.jpg"])
	assertFileBytes(t, filepath.Join(outputDir, "wasteland-knight.jpg"), sourceFiles["wasteland-knight.jpg"])
}

func TestImportCoupRoleImagesRemovesStaleCanonicalFilesForRole(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := t.TempDir()

	for _, name := range []string{"king.png", "blue-knight.jpg", "black-knight.jpg", "red-knight.jpg", "green-knight.jpg", "wasteland-knight.jpg"} {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(name), 0644); err != nil {
			t.Fatalf("write source %s: %v", name, err)
		}
	}
	for _, name := range []string{"king.jpg", "1001.jpg"} {
		if err := os.WriteFile(filepath.Join(outputDir, name), []byte("stale"), 0644); err != nil {
			t.Fatalf("write stale output %s: %v", name, err)
		}
	}

	if _, err := ImportCoupRoleImages(sourceDir, outputDir); err != nil {
		t.Fatalf("import Coup role images: %v", err)
	}

	for _, name := range []string{"king.jpg", "1001.jpg"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected stale %s to be removed, got err %v", name, err)
		}
	}
	assertFileBytes(t, filepath.Join(outputDir, "king.png"), []byte("king.png"))
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
