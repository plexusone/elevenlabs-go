package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/agentplexus/go-elevenlabs/ttsscript"
	"github.com/spf13/cobra"
)

var (
	ttsscriptLang     string
	ttsscriptOutput   string
	ttsscriptPerSlide bool
	ttsscriptManifest bool
	ttsscriptDryRun   bool
	ttsscriptModelID  string
)

var ttsscriptCmd = &cobra.Command{
	Use:   "ttsscript <script.json>",
	Short: "Generate TTS audio from a JSON script file",
	Long: `Generate TTS audio from a JSON script file using ElevenLabs.

The script file defines slides and segments with voice assignments,
pause timings, and multilingual text content.

Examples:
  # Generate with default settings
  elevenlabs ttsscript script.json

  # Specify language and output directory
  elevenlabs ttsscript -lang en -output ./audio script.json

  # Dry run to see what would be generated
  elevenlabs ttsscript -dry-run script.json

  # Generate per-slide concatenated audio (requires ffmpeg)
  elevenlabs ttsscript -per-slide script.json`,
	Args: cobra.ExactArgs(1),
	RunE: runTTSScript,
}

func init() {
	ttsscriptCmd.Flags().StringVarP(&ttsscriptLang, "lang", "l", "en", "Language code to generate")
	ttsscriptCmd.Flags().StringVarP(&ttsscriptOutput, "output", "o", "./output", "Output directory")
	ttsscriptCmd.Flags().BoolVar(&ttsscriptPerSlide, "per-slide", false, "Concatenate segments into per-slide audio files (requires ffmpeg)")
	ttsscriptCmd.Flags().BoolVar(&ttsscriptManifest, "manifest", true, "Generate manifest JSON file")
	ttsscriptCmd.Flags().BoolVar(&ttsscriptDryRun, "dry-run", false, "Show what would be generated without calling API")
	ttsscriptCmd.Flags().StringVarP(&ttsscriptModelID, "model", "m", "eleven_multilingual_v2", "ElevenLabs model ID")

	rootCmd.AddCommand(ttsscriptCmd)
}

func runTTSScript(cmd *cobra.Command, args []string) error {
	scriptPath := args[0]
	logger := slog.Default()

	// Check for API key (unless dry run)
	if !ttsscriptDryRun && os.Getenv("ELEVENLABS_API_KEY") == "" {
		return fmt.Errorf("ELEVENLABS_API_KEY environment variable is required")
	}

	// Check for ffmpeg if per-slide mode
	if ttsscriptPerSlide {
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			return fmt.Errorf("ffmpeg is required for --per-slide mode but was not found in PATH")
		}
	}

	// Load script
	script, err := ttsscript.LoadScript(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to load script: %w", err)
	}

	// Validate script
	if issues := script.Validate(); len(issues) > 0 {
		return fmt.Errorf("script validation failed:\n  - %s", strings.Join(issues, "\n  - "))
	}

	logger.Info("loaded script",
		"title", script.Title,
		"language", ttsscriptLang,
		"slides", script.SlideCount(),
		"segments", script.SegmentCount())

	// Compile script
	compiler := ttsscript.NewCompiler()
	segments, err := compiler.Compile(script, ttsscriptLang)
	if err != nil {
		return fmt.Errorf("failed to compile script: %w", err)
	}

	// Format for ElevenLabs
	formatter := ttsscript.NewElevenLabsFormatter()
	jobs := formatter.Format(segments)

	logger.Info("compiled script", "jobs", len(jobs))

	// Create output directory
	if err := os.MkdirAll(ttsscriptOutput, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate batch config
	config := ttsscript.NewBatchConfig(ttsscriptOutput)
	config.IncludeLanguageInFilename = true

	// Generate manifest
	manifestEntries := ttsscript.GenerateManifest(jobs, config, ttsscriptLang)

	if ttsscriptDryRun {
		logger.Info("dry run - would generate:")
		for _, entry := range manifestEntries {
			segType := "segment"
			if entry.IsTitleSegment {
				segType = "title"
			}
			logger.Info("job",
				"type", segType,
				"file", entry.OutputFile,
				"text", truncateText(entry.Text, 60),
				"voice", entry.VoiceID)
		}

		if ttsscriptPerSlide {
			logger.Info("per-slide output:")
			slideFiles := getSlideOutputFilesScript(manifestEntries, config, ttsscriptLang)
			for slide, file := range slideFiles {
				logger.Info("slide", "number", slide+1, "file", file)
			}
		}
		return nil
	}

	// Create ElevenLabs client
	client, err := elevenlabs.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create ElevenLabs client: %w", err)
	}

	ctx := context.Background()

	// Generate audio for each segment
	generatedFiles := make([]string, 0, len(jobs))
	for i, job := range jobs {
		if job.VoiceID == "" {
			logger.Warn("skipping segment: no voice ID configured", "segment", i+1)
			continue
		}

		outputFile := config.GenerateFilename(job, ttsscriptLang)

		segType := "segment"
		if job.IsTitleSegment {
			segType = "title"
		}

		logger.Info("generating",
			"progress", fmt.Sprintf("%d/%d", i+1, len(jobs)),
			"type", segType,
			"text", truncateText(job.Text, 50))

		resp, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
			VoiceID:       job.VoiceID,
			Text:          job.Text,
			ModelID:       ttsscriptModelID,
			VoiceSettings: elevenlabs.DefaultVoiceSettings(),
		})
		if err != nil {
			logger.Error("failed to generate speech", "error", err)
			continue
		}
		audio := resp.Audio

		f, err := os.Create(outputFile)
		if err != nil {
			logger.Error("failed to create file", "file", outputFile, "error", err)
			continue
		}

		_, err = io.Copy(f, audio)
		f.Close()
		if err != nil {
			logger.Error("failed to write file", "file", outputFile, "error", err)
			continue
		}

		logger.Info("saved", "file", outputFile)
		generatedFiles = append(generatedFiles, outputFile)
	}

	// Write manifest
	if ttsscriptManifest {
		manifestPath := filepath.Join(ttsscriptOutput, fmt.Sprintf("manifest_%s.json", ttsscriptLang))
		manifestData, err := json.MarshalIndent(manifestEntries, "", "  ")
		if err != nil {
			logger.Error("failed to marshal manifest", "error", err)
		} else if err := os.WriteFile(manifestPath, manifestData, 0600); err != nil {
			logger.Error("failed to write manifest", "error", err)
		} else {
			logger.Info("manifest saved", "file", manifestPath)
		}
	}

	// Concatenate per-slide if requested
	if ttsscriptPerSlide {
		logger.Info("concatenating per-slide audio")
		concatenatePerSlideScript(logger, manifestEntries, ttsscriptLang, ttsscriptOutput)
	}

	logger.Info("done", "generated", len(generatedFiles))
	return nil
}

// concatenatePerSlideScript uses ffmpeg to concatenate segment audio files into per-slide files.
func concatenatePerSlideScript(logger *slog.Logger, entries []ttsscript.ManifestEntry, language, outputDir string) {
	// Group entries by slide
	slideSegments := make(map[int][]ttsscript.ManifestEntry)
	for _, entry := range entries {
		slideSegments[entry.SlideIndex] = append(slideSegments[entry.SlideIndex], entry)
	}

	// Get sorted slide indices
	slideIndices := make([]int, 0, len(slideSegments))
	for idx := range slideSegments {
		slideIndices = append(slideIndices, idx)
	}
	sort.Ints(slideIndices)

	for _, slideIdx := range slideIndices {
		segments := slideSegments[slideIdx]

		// Sort segments: title first (SegmentIndex -1), then by segment index
		sort.Slice(segments, func(i, j int) bool {
			return segments[i].SegmentIndex < segments[j].SegmentIndex
		})

		// Skip if only one segment (no need to concatenate)
		if len(segments) == 1 {
			// Just copy/rename to slide output
			slideOutput := filepath.Join(outputDir, fmt.Sprintf("slide%02d_%s.mp3", slideIdx+1, language))
			if err := copyFileScript(segments[0].OutputFile, slideOutput); err != nil {
				logger.Error("failed to copy slide", "slide", slideIdx+1, "error", err)
				continue
			}
			logger.Info("slide copied", "slide", slideIdx+1, "file", slideOutput, "segments", 1)
			continue
		}

		// Create concat list file for ffmpeg
		listFile := filepath.Join(outputDir, fmt.Sprintf(".concat_slide%02d.txt", slideIdx+1))
		var listContent strings.Builder

		for i, seg := range segments {
			// Add pause before (as silence) if needed
			if seg.PauseBeforeMs > 0 && i > 0 {
				silenceFile, err := generateSilenceScript(outputDir, seg.PauseBeforeMs, slideIdx, i, "before")
				if err != nil {
					logger.Warn("failed to generate silence", "error", err)
				} else {
					fmt.Fprintf(&listContent, "file '%s'\n", filepath.Base(silenceFile))
				}
			}

			// Add the audio file
			fmt.Fprintf(&listContent, "file '%s'\n", filepath.Base(seg.OutputFile))

			// Add pause after (as silence) if needed
			if seg.PauseAfterMs > 0 {
				silenceFile, err := generateSilenceScript(outputDir, seg.PauseAfterMs, slideIdx, i, "after")
				if err != nil {
					logger.Warn("failed to generate silence", "error", err)
				} else {
					fmt.Fprintf(&listContent, "file '%s'\n", filepath.Base(silenceFile))
				}
			}
		}

		if err := os.WriteFile(listFile, []byte(listContent.String()), 0600); err != nil {
			logger.Error("failed to write concat list", "slide", slideIdx+1, "error", err)
			continue
		}

		// Run ffmpeg to concatenate
		slideOutput := filepath.Join(outputDir, fmt.Sprintf("slide%02d_%s.mp3", slideIdx+1, language))
		cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", slideOutput)
		cmd.Dir = outputDir
		if output, err := cmd.CombinedOutput(); err != nil {
			logger.Error("ffmpeg failed", "slide", slideIdx+1, "error", err, "output", string(output))
			continue
		}

		// Clean up temp files
		os.Remove(listFile)
		cleanupSilenceFilesScript(outputDir, slideIdx)

		logger.Info("slide concatenated", "slide", slideIdx+1, "file", slideOutput, "segments", len(segments))
	}
}

// generateSilenceScript creates a silent audio file of the specified duration.
func generateSilenceScript(outputDir string, durationMs, slideIdx, segIdx int, position string) (string, error) {
	filename := filepath.Join(outputDir, fmt.Sprintf(".silence_s%02d_%02d_%s.mp3", slideIdx, segIdx, position))
	duration := float64(durationMs) / 1000.0

	// #nosec G204 -- filename is constructed from user-controlled outputDir flag, which is intentional for CLI tools
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		fmt.Sprintf("anullsrc=r=44100:cl=mono:d=%.3f", duration),
		"-c:a", "libmp3lame", "-q:a", "9", filename)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg silence generation failed: %v\n%s", err, string(output))
	}

	return filename, nil
}

// cleanupSilenceFilesScript removes temporary silence files for a slide.
func cleanupSilenceFilesScript(outputDir string, slideIdx int) {
	pattern := filepath.Join(outputDir, fmt.Sprintf(".silence_s%02d_*.mp3", slideIdx))
	files, _ := filepath.Glob(pattern)
	for _, f := range files {
		os.Remove(f)
	}
}

// copyFileScript copies a file from src to dst.
func copyFileScript(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// getSlideOutputFilesScript returns a map of slide index to output file path.
func getSlideOutputFilesScript(entries []ttsscript.ManifestEntry, config *ttsscript.BatchConfig, language string) map[int]string {
	slides := make(map[int]string)
	for _, entry := range entries {
		if _, exists := slides[entry.SlideIndex]; !exists {
			slides[entry.SlideIndex] = filepath.Join(config.OutputDir, fmt.Sprintf("slide%02d_%s.mp3", entry.SlideIndex+1, language))
		}
	}
	return slides
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
