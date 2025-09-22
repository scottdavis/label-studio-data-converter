package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// createTestFiles creates test files for testing
func createTestFiles(t testing.TB, baseDir string) {
	// Create directories
	dirs := []string{
		filepath.Join(baseDir, "images"),
		filepath.Join(baseDir, "labels"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test image files (empty files for testing)
	imageFiles := []string{"image1.jpg", "image2.png", "image3.jpeg"}
	for _, file := range imageFiles {
		path := filepath.Join(baseDir, "images", file)
		if err := os.WriteFile(path, []byte("fake image data"), 0644); err != nil {
			t.Fatalf("Failed to create test image %s: %v", path, err)
		}
	}

	// Create test label files with YOLO format
	labelData := map[string]string{
		"image1.txt": "0 0.5 0.5 0.3 0.3\n1 0.2 0.8 0.1 0.1\n",
		"image2.txt": "0 0.4 0.6 0.2 0.4\n",
		"image3.txt": "1 0.7 0.3 0.3 0.2\n0 0.1 0.9 0.1 0.1\n",
	}

	for file, content := range labelData {
		path := filepath.Join(baseDir, "labels", file)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test label %s: %v", path, err)
		}
	}

	// Create classes.txt
	classesPath := filepath.Join(baseDir, "classes.txt")
	classesContent := "book\nperson\n"
	if err := os.WriteFile(classesPath, []byte(classesContent), 0644); err != nil {
		t.Fatalf("Failed to create classes.txt: %v", err)
	}

	// Create notes.json (optional)
	notesPath := filepath.Join(baseDir, "notes.json")
	notesContent := `{
		"categories": [
			{"id": 0, "name": "book"},
			{"id": 1, "name": "person"}
		],
		"info": {
			"year": 2025,
			"version": "1.0",
			"contributor": "Label Studio"
		}
	}`
	if err := os.WriteFile(notesPath, []byte(notesContent), 0644); err != nil {
		t.Fatalf("Failed to create notes.json: %v", err)
	}
}

func TestNewConverter(t *testing.T) {
	config := Config{
		SourceDir:  "/test/source",
		OutputDir:  "/test/output",
		TrainSplit: 0.8,
		Seed:       42,
	}

	converter := NewConverter(config)

	if converter.config != config {
		t.Errorf("Expected config %+v, got %+v", config, converter.config)
	}
}

func TestValidateSourceStructure(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()

	config := Config{SourceDir: tempDir}
	converter := NewConverter(config)

	// Test with missing directories
	err := converter.ValidateSourceStructure()
	if err == nil {
		t.Error("Expected error for missing directories, got nil")
	}

	// Create test files
	createTestFiles(t, tempDir)

	// Test with valid structure
	err = converter.ValidateSourceStructure()
	if err != nil {
		t.Errorf("Expected no error for valid structure, got: %v", err)
	}
}

func TestLoadClasses(t *testing.T) {
	tempDir := t.TempDir()
	createTestFiles(t, tempDir)

	config := Config{SourceDir: tempDir}
	converter := NewConverter(config)

	classes, err := converter.LoadClasses()
	if err != nil {
		t.Fatalf("Failed to load classes: %v", err)
	}

	expected := []string{"book", "person"}
	if !reflect.DeepEqual(classes, expected) {
		t.Errorf("Expected classes %v, got %v", expected, classes)
	}
}

func TestLoadClassesFileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	config := Config{SourceDir: tempDir}
	converter := NewConverter(config)

	_, err := converter.LoadClasses()
	if err == nil {
		t.Error("Expected error for missing classes.txt, got nil")
	}

	if !strings.Contains(err.Error(), "classes.txt") {
		t.Errorf("Expected error message to contain 'classes.txt', got: %v", err)
	}
}

func TestGetImageLabelPairs(t *testing.T) {
	tempDir := t.TempDir()
	createTestFiles(t, tempDir)

	config := Config{SourceDir: tempDir}
	converter := NewConverter(config)

	pairs, err := converter.GetImageLabelPairs()
	if err != nil {
		t.Fatalf("Failed to get image-label pairs: %v", err)
	}

	if len(pairs) != 3 {
		t.Errorf("Expected 3 pairs, got %d", len(pairs))
	}

	// Check that all pairs have corresponding files
	for _, pair := range pairs {
		if _, err := os.Stat(pair.ImagePath); os.IsNotExist(err) {
			t.Errorf("Image file does not exist: %s", pair.ImagePath)
		}
		if _, err := os.Stat(pair.LabelPath); os.IsNotExist(err) {
			t.Errorf("Label file does not exist: %s", pair.LabelPath)
		}
	}
}

func TestSplitDataset(t *testing.T) {
	tempDir := t.TempDir()
	createTestFiles(t, tempDir)

	config := Config{
		SourceDir:  tempDir,
		TrainSplit: 0.8,
		Seed:       42,
	}
	converter := NewConverter(config)

	pairs, err := converter.GetImageLabelPairs()
	if err != nil {
		t.Fatalf("Failed to get pairs: %v", err)
	}

	trainPairs, valPairs := converter.SplitDataset(pairs)

	expectedTrainCount := int(float64(len(pairs)) * 0.8)
	expectedValCount := len(pairs) - expectedTrainCount

	if len(trainPairs) != expectedTrainCount {
		t.Errorf("Expected %d training pairs, got %d", expectedTrainCount, len(trainPairs))
	}

	if len(valPairs) != expectedValCount {
		t.Errorf("Expected %d validation pairs, got %d", expectedValCount, len(valPairs))
	}

	// Test that total count is preserved
	if len(trainPairs)+len(valPairs) != len(pairs) {
		t.Error("Total count not preserved after split")
	}
}

func TestCreateYOLOStructure(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "yolo_output")

	config := Config{OutputDir: outputDir}
	converter := NewConverter(config)

	err := converter.CreateYOLOStructure()
	if err != nil {
		t.Fatalf("Failed to create YOLO structure: %v", err)
	}

	// Check that all required directories exist
	requiredDirs := []string{
		filepath.Join(outputDir, "images", "train"),
		filepath.Join(outputDir, "images", "val"),
		filepath.Join(outputDir, "labels", "train"),
		filepath.Join(outputDir, "labels", "val"),
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Required directory not created: %s", dir)
		}
	}
}

func TestValidateLabels(t *testing.T) {
	tempDir := t.TempDir()
	createTestFiles(t, tempDir)

	config := Config{SourceDir: tempDir}
	converter := NewConverter(config)

	pairs, err := converter.GetImageLabelPairs()
	if err != nil {
		t.Fatalf("Failed to get pairs: %v", err)
	}

	stats, err := converter.ValidateLabels(pairs)
	if err != nil {
		t.Fatalf("Failed to validate labels: %v", err)
	}

	if stats.TotalFiles != len(pairs) {
		t.Errorf("Expected total files %d, got %d", len(pairs), stats.TotalFiles)
	}

	// We created 5 annotations total (2 + 1 + 2)
	expectedAnnotations := 5
	if stats.TotalAnnotations != expectedAnnotations {
		t.Errorf("Expected %d annotations, got %d", expectedAnnotations, stats.TotalAnnotations)
	}

	if stats.FilesWithAnnotations != 3 {
		t.Errorf("Expected 3 files with annotations, got %d", stats.FilesWithAnnotations)
	}
}

func TestValidateLabelsInvalidFormat(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal structure
	dirs := []string{
		filepath.Join(tempDir, "images"),
		filepath.Join(tempDir, "labels"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Create image
	imagePath := filepath.Join(tempDir, "images", "test.jpg")
	if err := os.WriteFile(imagePath, []byte("fake"), 0644); err != nil {
		t.Fatalf("Failed to create image: %v", err)
	}

	// Create invalid label file
	labelPath := filepath.Join(tempDir, "labels", "test.txt")
	invalidContent := "0 0.5 0.5\n1 invalid 0.8 0.1 0.1\n"
	if err := os.WriteFile(labelPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	// Create classes.txt
	classesPath := filepath.Join(tempDir, "classes.txt")
	if err := os.WriteFile(classesPath, []byte("test\n"), 0644); err != nil {
		t.Fatalf("Failed to create classes.txt: %v", err)
	}

	config := Config{SourceDir: tempDir}
	converter := NewConverter(config)

	pairs, err := converter.GetImageLabelPairs()
	if err != nil {
		t.Fatalf("Failed to get pairs: %v", err)
	}

	stats, err := converter.ValidateLabels(pairs)
	if err != nil {
		t.Fatalf("Failed to validate labels: %v", err)
	}

	// Should detect invalid lines
	if stats.InvalidLines == 0 {
		t.Error("Expected invalid lines to be detected")
	}
}

func TestCopyFiles(t *testing.T) {
	tempDir := t.TempDir()
	createTestFiles(t, tempDir)
	outputDir := filepath.Join(tempDir, "output")

	config := Config{
		SourceDir: tempDir,
		OutputDir: outputDir,
	}
	converter := NewConverter(config)

	// Create YOLO structure first
	err := converter.CreateYOLOStructure()
	if err != nil {
		t.Fatalf("Failed to create YOLO structure: %v", err)
	}

	pairs, err := converter.GetImageLabelPairs()
	if err != nil {
		t.Fatalf("Failed to get pairs: %v", err)
	}

	// Test copying to train directory
	err = converter.CopyFiles(pairs, "train")
	if err != nil {
		t.Fatalf("Failed to copy files: %v", err)
	}

	// Check that files were copied
	trainImagesDir := filepath.Join(outputDir, "images", "train")
	trainLabelsDir := filepath.Join(outputDir, "labels", "train")

	for _, pair := range pairs {
		imageName := filepath.Base(pair.ImagePath)
		labelName := filepath.Base(pair.LabelPath)

		copiedImagePath := filepath.Join(trainImagesDir, imageName)
		copiedLabelPath := filepath.Join(trainLabelsDir, labelName)

		if _, err := os.Stat(copiedImagePath); os.IsNotExist(err) {
			t.Errorf("Image file not copied: %s", copiedImagePath)
		}

		if _, err := os.Stat(copiedLabelPath); os.IsNotExist(err) {
			t.Errorf("Label file not copied: %s", copiedLabelPath)
		}
	}
}

func TestCreateYAMLConfig(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "output")

	config := Config{OutputDir: outputDir}
	converter := NewConverter(config)

	// Create output directory
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	classes := []string{"book", "person"}
	err = converter.CreateYAMLConfig(classes)
	if err != nil {
		t.Fatalf("Failed to create YAML config: %v", err)
	}

	// Check that YAML file exists
	yamlPath := filepath.Join(outputDir, "data.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Error("YAML config file not created")
	}

	// Read and verify content
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("Failed to read YAML file: %v", err)
	}

	yamlContent := string(content)
	if !strings.Contains(yamlContent, "nc: 2") {
		t.Error("YAML should contain correct number of classes")
	}

	if !strings.Contains(yamlContent, "book") {
		t.Error("YAML should contain class names")
	}
}

func TestFullConversion(t *testing.T) {
	tempDir := t.TempDir()
	createTestFiles(t, tempDir)
	outputDir := filepath.Join(tempDir, "yolo_output")

	config := Config{
		SourceDir:  tempDir,
		OutputDir:  outputDir,
		TrainSplit: 0.8,
		Seed:       42,
	}

	converter := NewConverter(config)

	err := converter.Convert()
	if err != nil {
		t.Fatalf("Full conversion failed: %v", err)
	}

	// Verify output structure
	requiredPaths := []string{
		filepath.Join(outputDir, "data.yaml"),
		filepath.Join(outputDir, "images", "train"),
		filepath.Join(outputDir, "images", "val"),
		filepath.Join(outputDir, "labels", "train"),
		filepath.Join(outputDir, "labels", "val"),
	}

	for _, path := range requiredPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Required path not found: %s", path)
		}
	}

	// Check that files were distributed
	trainImagesDir := filepath.Join(outputDir, "images", "train")
	valImagesDir := filepath.Join(outputDir, "images", "val")

	trainFiles, err := os.ReadDir(trainImagesDir)
	if err != nil {
		t.Fatalf("Failed to read train images directory: %v", err)
	}

	valFiles, err := os.ReadDir(valImagesDir)
	if err != nil {
		t.Fatalf("Failed to read val images directory: %v", err)
	}

	totalFiles := len(trainFiles) + len(valFiles)
	if totalFiles != 3 { // We created 3 image files
		t.Errorf("Expected 3 total files, got %d", totalFiles)
	}
}

// Benchmark tests
func BenchmarkSplitDataset(b *testing.B) {
	// Create test pairs
	pairs := make([]LabelPair, 1000)
	for i := 0; i < 1000; i++ {
		pairs[i] = LabelPair{
			ImagePath: "image.jpg",
			LabelPath: "label.txt",
		}
	}

	config := Config{TrainSplit: 0.8, Seed: 42}
	converter := NewConverter(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		converter.SplitDataset(pairs)
	}
}

func BenchmarkValidateLabels(b *testing.B) {
	tempDir := b.TempDir()
	createTestFiles(b, tempDir)

	config := Config{SourceDir: tempDir}
	converter := NewConverter(config)

	pairs, err := converter.GetImageLabelPairs()
	if err != nil {
		b.Fatalf("Failed to get pairs: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		converter.ValidateLabels(pairs)
	}
}
