package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"treacherest/internal/game"
)

const defaultOutputDir = "static/images/coup"

// ImportResult summarizes copied Coup role images.
type ImportResult struct {
	Imported int
	Files    []string
}

func main() {
	sourceDir := flag.String("source", "", "directory containing user-provided Coup role images")
	outputDir := flag.String("output", defaultOutputDir, "directory to write canonical embedded Coup role images")
	flag.Parse()

	if *sourceDir == "" {
		fmt.Fprintln(os.Stderr, "missing -source")
		os.Exit(2)
	}

	result, err := ImportCoupRoleImages(*sourceDir, *outputDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, file := range result.Files {
		fmt.Printf("Imported %s\n", file)
	}
	fmt.Printf("Imported %d Coup role images into %s\n", result.Imported, *outputDir)
}

// ImportCoupRoleImages copies supported role image files into canonical slug-based filenames.
func ImportCoupRoleImages(sourceDir, outputDir string) (ImportResult, error) {
	var result ImportResult

	if sourceDir == "" {
		return result, errors.New("source directory is required")
	}
	if outputDir == "" {
		outputDir = defaultOutputDir
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return result, fmt.Errorf("create output directory: %w", err)
	}

	for _, roleImage := range game.CoupRoleImageManifest() {
		sourcePath, extension, err := findSourceImage(sourceDir, roleImage)
		if err != nil {
			return result, err
		}

		if err := removeExistingCanonicalImages(outputDir, roleImage); err != nil {
			return result, err
		}

		outputPath := filepath.Join(outputDir, roleImage.Slug+extension)
		if err := copyFile(sourcePath, outputPath); err != nil {
			return result, fmt.Errorf("copy %s to %s: %w", sourcePath, outputPath, err)
		}
		result.Imported++
		result.Files = append(result.Files, outputPath)
	}

	return result, nil
}

func findSourceImage(sourceDir string, roleImage game.CoupRoleImage) (string, string, error) {
	for _, baseName := range sourceBaseNames(roleImage) {
		for _, extension := range game.SupportedCoupRoleImageExtensions() {
			path := filepath.Join(sourceDir, baseName+extension)
			if _, err := os.Stat(path); err == nil {
				return path, extension, nil
			} else if !os.IsNotExist(err) {
				return "", "", fmt.Errorf("stat %s: %w", path, err)
			}
		}
	}
	return "", "", fmt.Errorf("missing image for %s; expected one of %s with extension %s", roleImage.Role, strings.Join(sourceBaseNames(roleImage), ", "), strings.Join(game.SupportedCoupRoleImageExtensions(), ", "))
}

func sourceBaseNames(roleImage game.CoupRoleImage) []string {
	roleName := strings.ToLower(strings.ReplaceAll(string(roleImage.Role), " ", "-"))
	return []string{
		roleImage.Slug,
		roleName,
		fmt.Sprintf("%d", roleImage.ID),
	}
}

func removeExistingCanonicalImages(outputDir string, roleImage game.CoupRoleImage) error {
	for _, extension := range game.SupportedCoupRoleImageExtensions() {
		for _, baseName := range canonicalBaseNames(roleImage) {
			path := filepath.Join(outputDir, baseName+extension)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove stale image %s: %w", path, err)
			}
		}
	}
	return nil
}

func canonicalBaseNames(roleImage game.CoupRoleImage) []string {
	return []string{
		roleImage.Slug,
		fmt.Sprintf("%d", roleImage.ID),
	}
}

func copyFile(sourcePath, outputPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, source)
	return err
}
