package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Version information (set during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

// Config holds the conversion configuration
type Config struct {
	SourceDir  string
	OutputDir  string
	TrainSplit float64
	Seed       int64
}

// LabelPair represents an image-label file pair
type LabelPair struct {
	ImagePath string
	LabelPath string
}

// ValidationStats holds statistics about label validation
type ValidationStats struct {
	TotalFiles           int `json:"total_files"`
	TotalAnnotations     int `json:"total_annotations"`
	FilesWithAnnotations int `json:"files_with_annotations"`
	EmptyFiles           int `json:"empty_files"`
	InvalidLines         int `json:"invalid_lines"`
}

// YAMLConfig represents the YOLO dataset configuration
type YAMLConfig struct {
	Path  string   `yaml:"path"`
	Train string   `yaml:"train"`
	Val   string   `yaml:"val"`
	NC    int      `yaml:"nc"`
	Names []string `yaml:"names"`
}

// NotesInfo represents the structure of notes.json from Label Studio
type NotesInfo struct {
	Categories []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"categories"`
	Info struct {
		Year        int    `json:"year"`
		Version     string `json:"version"`
		Contributor string `json:"contributor"`
	} `json:"info"`
}

// Converter handles the Label Studio to YOLO conversion
type Converter struct {
	config Config
}

// NewConverter creates a new converter instance
func NewConverter(config Config) *Converter {
	return &Converter{config: config}
}

// ValidateSourceStructure checks if the source directory has the expected structure
func (c *Converter) ValidateSourceStructure() error {
	requiredDirs := []string{
		filepath.Join(c.config.SourceDir, "images"),
		filepath.Join(c.config.SourceDir, "labels"),
	}

	requiredFiles := []string{
		filepath.Join(c.config.SourceDir, "classes.txt"),
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("required directory not found: %s", dir)
		}
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("required file not found: %s", file)
		}
	}

	return nil
}

// LoadClasses loads class names from classes.txt
func (c *Converter) LoadClasses() ([]string, error) {
	classesPath := filepath.Join(c.config.SourceDir, "classes.txt")
	file, err := os.Open(classesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open classes.txt: %w", err)
	}
	defer file.Close()

	var classes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			classes = append(classes, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading classes.txt: %w", err)
	}

	fmt.Printf("Found %d classes: %v\n", len(classes), classes)
	return classes, nil
}

// GetImageLabelPairs finds matching image and label file pairs
func (c *Converter) GetImageLabelPairs() ([]LabelPair, error) {
	imagesDir := filepath.Join(c.config.SourceDir, "images")
	labelsDir := filepath.Join(c.config.SourceDir, "labels")

	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".tiff": true,
		".webp": true,
	}

	var pairs []LabelPair

	err := filepath.Walk(imagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if !imageExtensions[ext] {
			return nil
		}

		// Find corresponding label file
		baseName := strings.TrimSuffix(info.Name(), ext)
		labelPath := filepath.Join(labelsDir, baseName+".txt")

		if _, err := os.Stat(labelPath); err == nil {
			pairs = append(pairs, LabelPair{
				ImagePath: path,
				LabelPath: labelPath,
			})
		} else {
			fmt.Printf("Warning: No label file found for %s\n", info.Name())
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning images directory: %w", err)
	}

	fmt.Printf("Found %d image-label pairs\n", len(pairs))
	return pairs, nil
}

// SplitDataset splits the dataset into train and validation sets
func (c *Converter) SplitDataset(pairs []LabelPair) ([]LabelPair, []LabelPair) {
	// Set random seed for reproducible splits
	rand.Seed(c.config.Seed)

	// Shuffle the pairs
	shuffled := make([]LabelPair, len(pairs))
	copy(shuffled, pairs)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Calculate split index
	trainCount := int(float64(len(shuffled)) * c.config.TrainSplit)

	trainPairs := shuffled[:trainCount]
	valPairs := shuffled[trainCount:]

	fmt.Printf("Dataset split: %d training, %d validation\n", len(trainPairs), len(valPairs))
	return trainPairs, valPairs
}

// CreateYOLOStructure creates the YOLO directory structure
func (c *Converter) CreateYOLOStructure() error {
	dirsToCreate := []string{
		filepath.Join(c.config.OutputDir, "images", "train"),
		filepath.Join(c.config.OutputDir, "images", "val"),
		filepath.Join(c.config.OutputDir, "labels", "train"),
		filepath.Join(c.config.OutputDir, "labels", "val"),
	}

	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	fmt.Printf("Created YOLO directory structure in: %s\n", c.config.OutputDir)
	return nil
}

// CopyFiles copies image and label files to the appropriate YOLO directories
func (c *Converter) CopyFiles(pairs []LabelPair, splitType string) error {
	imagesDestDir := filepath.Join(c.config.OutputDir, "images", splitType)
	labelsDestDir := filepath.Join(c.config.OutputDir, "labels", splitType)

	for _, pair := range pairs {
		// Copy image
		imageName := filepath.Base(pair.ImagePath)
		imageDest := filepath.Join(imagesDestDir, imageName)
		if err := copyFile(pair.ImagePath, imageDest); err != nil {
			return fmt.Errorf("failed to copy image %s: %w", pair.ImagePath, err)
		}

		// Copy label
		labelName := filepath.Base(pair.LabelPath)
		labelDest := filepath.Join(labelsDestDir, labelName)
		if err := copyFile(pair.LabelPath, labelDest); err != nil {
			return fmt.Errorf("failed to copy label %s: %w", pair.LabelPath, err)
		}
	}

	fmt.Printf("Copied %d %s files\n", len(pairs), splitType)
	return nil
}

// CreateYAMLConfig creates the YAML configuration file for YOLO
func (c *Converter) CreateYAMLConfig(classes []string) error {
	absOutputDir, err := filepath.Abs(c.config.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	config := YAMLConfig{
		Path:  absOutputDir,
		Train: "images/train",
		Val:   "images/val",
		NC:    len(classes),
		Names: classes,
	}

	yamlPath := filepath.Join(c.config.OutputDir, "data.yaml")
	file, err := os.Create(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %w", err)
	}
	defer file.Close()

	// Write header comment
	header := fmt.Sprintf("# YOLO Dataset Configuration\n# Generated from Label Studio export\n# Generated at: %s\n\n", time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write YAML header: %w", err)
	}

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	defer encoder.Close()

	if err := encoder.Encode(&config); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	fmt.Printf("Created YAML config: %s\n", yamlPath)
	return nil
}

// ValidateLabels validates label files and counts annotations
func (c *Converter) ValidateLabels(pairs []LabelPair) (*ValidationStats, error) {
	stats := &ValidationStats{
		TotalFiles: len(pairs),
	}

	for _, pair := range pairs {
		file, err := os.Open(pair.LabelPath)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", pair.LabelPath, err)
			stats.InvalidLines++
			continue
		}

		validLines := 0
		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) != 5 {
				fmt.Printf("Warning: Wrong number of values in %s:%d\n", filepath.Base(pair.LabelPath), lineNum)
				stats.InvalidLines++
				continue
			}

			// Validate format: class_id x_center y_center width height
			if _, err := strconv.Atoi(parts[0]); err != nil {
				fmt.Printf("Warning: Invalid class_id in %s:%d\n", filepath.Base(pair.LabelPath), lineNum)
				stats.InvalidLines++
				continue
			}

			allValid := true
			for i := 1; i < 5; i++ {
				coord, err := strconv.ParseFloat(parts[i], 64)
				if err != nil {
					fmt.Printf("Warning: Invalid coordinate in %s:%d\n", filepath.Base(pair.LabelPath), lineNum)
					stats.InvalidLines++
					allValid = false
					break
				}
				if coord < 0 || coord > 1 {
					fmt.Printf("Warning: Non-normalized coordinates in %s:%d\n", filepath.Base(pair.LabelPath), lineNum)
					stats.InvalidLines++
					allValid = false
					break
				}
			}

			if allValid {
				validLines++
			}
		}

		file.Close()

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error scanning %s: %v\n", pair.LabelPath, err)
			stats.InvalidLines++
			continue
		}

		stats.TotalAnnotations += validLines
		if validLines > 0 {
			stats.FilesWithAnnotations++
		} else {
			stats.EmptyFiles++
		}
	}

	return stats, nil
}

// Convert performs the main conversion process
func (c *Converter) Convert() error {
	fmt.Println("Starting Label Studio to YOLO conversion...")
	fmt.Printf("Source: %s\n", c.config.SourceDir)
	fmt.Printf("Output: %s\n", c.config.OutputDir)
	fmt.Printf("Train split: %.1f%%\n", c.config.TrainSplit*100)

	// Validate source structure
	if err := c.ValidateSourceStructure(); err != nil {
		return err
	}

	// Load classes
	classes, err := c.LoadClasses()
	if err != nil {
		return err
	}

	// Get image-label pairs
	pairs, err := c.GetImageLabelPairs()
	if err != nil {
		return err
	}

	if len(pairs) == 0 {
		return fmt.Errorf("no valid image-label pairs found")
	}

	// Validate labels
	fmt.Println("\nValidating labels...")
	stats, err := c.ValidateLabels(pairs)
	if err != nil {
		return err
	}
	fmt.Printf("Validation stats: %+v\n", stats)

	// Split dataset
	trainPairs, valPairs := c.SplitDataset(pairs)

	// Create YOLO structure
	if err := c.CreateYOLOStructure(); err != nil {
		return err
	}

	// Copy files
	if err := c.CopyFiles(trainPairs, "train"); err != nil {
		return err
	}
	if err := c.CopyFiles(valPairs, "val"); err != nil {
		return err
	}

	// Create YAML config
	if err := c.CreateYAMLConfig(classes); err != nil {
		return err
	}

	fmt.Println("\nConversion completed successfully!")
	fmt.Printf("Dataset ready for YOLO training at: %s\n", c.config.OutputDir)
	fmt.Printf("Training images: %d\n", len(trainPairs))
	fmt.Printf("Validation images: %d\n", len(valPairs))
	fmt.Printf("Total annotations: %d\n", stats.TotalAnnotations)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

func main() {
	var config Config

	flag.StringVar(&config.SourceDir, "source", ".", "Path to Label Studio export directory")
	flag.StringVar(&config.OutputDir, "output", "./yolo_dataset", "Path where YOLO dataset will be created")
	flag.Float64Var(&config.TrainSplit, "train-split", 0.8, "Fraction of data for training (default: 0.8)")
	flag.Int64Var(&config.Seed, "seed", 42, "Random seed for reproducible splits (default: 42)")

	var showHelp bool
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message")

	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information")

	flag.Parse()

	if showVersion {
		fmt.Printf("labelstudio-to-yolo version %s\n", Version)
		fmt.Printf("Built: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Go version: %s\n", strings.TrimPrefix(runtime.Version(), "go"))
		return
	}

	if showHelp {
		fmt.Println("Label Studio to YOLO Converter")
		fmt.Println("===============================")
		fmt.Println()
		fmt.Println("Converts Label Studio export to YOLO training data format.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [flags]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Printf("  %s -source . -output ./yolo_dataset\n", os.Args[0])
		fmt.Printf("  %s -source /path/to/labelstudio -output /path/to/yolo -train-split 0.7\n", os.Args[0])
		return
	}

	converter := NewConverter(config)
	if err := converter.Convert(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
