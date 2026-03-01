// Example: Using ttsscript to generate multilingual TTS audio
//
// This example shows how to:
// 1. Load a script from JSON
// 2. Compile it for a specific language
// 3. Generate audio using ElevenLabs
// 4. Export to SSML for other TTS engines
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	elevenlabs "github.com/plexusone/elevenlabs-go"
	"github.com/plexusone/elevenlabs-go/ttsscript"
)

func main() {
	// Example 1: Create a script programmatically
	script := createExampleScript()

	// Validate the script
	if issues := script.Validate(); len(issues) > 0 {
		log.Fatalf("Script validation failed: %v", issues)
	}

	fmt.Printf("Script: %s\n", script.Title)
	fmt.Printf("Languages: %v\n", script.Languages())
	fmt.Printf("Slides: %d, Segments: %d\n", script.SlideCount(), script.SegmentCount())

	// Example 2: Compile for ElevenLabs
	fmt.Println("\n--- ElevenLabs Output ---")
	compileForElevenLabs(script, "en")

	// Example 3: Compile to SSML
	fmt.Println("\n--- SSML Output ---")
	compileToSSML(script, "en")

	// Example 4: Generate audio (requires API key)
	if os.Getenv("ELEVENLABS_API_KEY") != "" {
		fmt.Println("\n--- Generating Audio ---")
		generateAudio(script, "en")
	} else {
		fmt.Println("\n(Set ELEVENLABS_API_KEY to generate audio)")
	}

	// Example 5: Save script to JSON
	if err := script.Save("output_script.json"); err != nil {
		log.Printf("Failed to save script: %v", err)
	} else {
		fmt.Println("\nScript saved to output_script.json")
	}
}

func createExampleScript() *ttsscript.Script {
	return &ttsscript.Script{
		Title:           "Introduction to Go",
		Description:     "A multilingual introduction to the Go programming language",
		DefaultLanguage: "en",
		DefaultVoices: map[string]string{
			"en": "21m00Tcm4TlvDq8ikWAM", // Rachel
			"es": "EXAVITQu4vr4xnSDxMaL", // Bella
		},
		Pronunciations: map[string]map[string]string{
			"API": {"en": "A P I", "es": "A P I"},
			"SDK": {"en": "S D K", "es": "S D K"},
			"Go":  {"en": "Go", "es": "Go"},
		},
		Slides: []ttsscript.Slide{
			{
				Title: "Welcome",
				Segments: []ttsscript.Segment{
					{
						Text: map[string]string{
							"en": "Welcome to this introduction to the Go programming language.",
							"es": "Bienvenidos a esta introducción al lenguaje de programación Go.",
						},
						PauseAfter: "800ms",
					},
					{
						Text: map[string]string{
							"en": "Go is a fast, simple, and powerful language created by Google.",
							"es": "Go es un lenguaje rápido, simple y poderoso creado por Google.",
						},
						PauseAfter: "500ms",
					},
				},
			},
			{
				Title: "Key Features",
				Segments: []ttsscript.Segment{
					{
						Text: map[string]string{
							"en": "Go has several key features.",
							"es": "Go tiene varias características clave.",
						},
						PauseAfter: "300ms",
					},
					{
						Text: map[string]string{
							"en": "First, it compiles directly to machine code.",
							"es": "Primero, compila directamente a código máquina.",
						},
						PauseAfter: "300ms",
					},
					{
						Text: map[string]string{
							"en": "Second, it has excellent support for concurrent programming.",
							"es": "Segundo, tiene excelente soporte para programación concurrente.",
						},
						PauseAfter: "300ms",
					},
					{
						Text: map[string]string{
							"en": "And finally, it includes a comprehensive standard library.",
							"es": "Y finalmente, incluye una biblioteca estándar completa.",
						},
						Emphasis: "moderate",
					},
				},
			},
		},
	}
}

func compileForElevenLabs(script *ttsscript.Script, language string) {
	compiler := ttsscript.NewCompiler()

	// Add additional pronunciations if needed
	compiler.AddPronunciation("goroutine", "en", "go routine")
	compiler.AddPronunciation("goroutine", "es", "go rutina")

	segments, err := compiler.Compile(script, language)
	if err != nil {
		log.Fatalf("Compile failed: %v", err)
	}

	formatter := ttsscript.NewElevenLabsFormatter()
	jobs := formatter.Format(segments)

	fmt.Printf("Generated %d TTS jobs for language: %s\n", len(jobs), language)
	for i, job := range jobs {
		fmt.Printf("  %d. [%s] Voice: %s\n", i+1, job.SlideTitle, job.VoiceID)
		fmt.Printf("     Text: %s\n", truncate(job.Text, 60))
		if job.PauseAfterMs > 0 {
			fmt.Printf("     Pause after: %dms\n", job.PauseAfterMs)
		}
	}

	// Generate manifest for batch processing
	config := ttsscript.NewBatchConfig("./output")
	manifest := ttsscript.GenerateManifest(jobs, config, language)

	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	fmt.Printf("\nManifest:\n%s\n", string(manifestJSON))
}

func compileToSSML(script *ttsscript.Script, language string) {
	formatter := ttsscript.NewSSMLFormatter()
	ssml, err := formatter.FormatScript(script, language)
	if err != nil {
		log.Fatalf("SSML formatting failed: %v", err)
	}

	fmt.Println(ssml)
}

func generateAudio(script *ttsscript.Script, language string) {
	client, err := elevenlabs.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	compiler := ttsscript.NewCompiler()
	segments, err := compiler.Compile(script, language)
	if err != nil {
		log.Fatalf("Compile failed: %v", err)
	}

	formatter := ttsscript.NewElevenLabsFormatter()
	jobs := formatter.Format(segments)

	ctx := context.Background()

	// Create output directory
	if err := os.MkdirAll("./output", 0750); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	for i, job := range jobs {
		if job.VoiceID == "" {
			log.Printf("Skipping segment %d: no voice ID", i+1)
			continue
		}

		fmt.Printf("Generating segment %d/%d: %s\n", i+1, len(jobs), truncate(job.Text, 40))

		audio, err := client.TextToSpeech().Simple(ctx, job.VoiceID, job.Text)
		if err != nil {
			log.Printf("Failed to generate segment %d: %v", i+1, err)
			continue
		}

		filename := fmt.Sprintf("./output/segment_%02d_%s.mp3", i+1, language)
		f, err := os.Create(filename)
		if err != nil {
			log.Printf("Failed to create file: %v", err)
			continue
		}

		_, err = io.Copy(f, audio)
		f.Close()
		if err != nil {
			log.Printf("Failed to write file: %v", err)
			continue
		}

		fmt.Printf("  Saved: %s\n", filename)
	}

	fmt.Println("Audio generation complete!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
