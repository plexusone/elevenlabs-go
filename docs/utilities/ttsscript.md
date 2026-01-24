# TTS Script Package

The `ttsscript` package provides a structured format for authoring multilingual TTS scripts.

## Installation

The package is included with go-elevenlabs:

```go
import "github.com/agentplexus/go-elevenlabs/ttsscript"
```

## Package Overview

```
ttsscript/
├── script.go      # Script, Slide, Segment types
├── compiler.go    # Compiler with pronunciation handling
├── ssml.go        # SSML formatter
├── elevenlabs.go  # ElevenLabs formatter
└── doc.go         # Package documentation
```

## Types

### Script

```go
type Script struct {
    Title           string
    Description     string
    DefaultLanguage string
    DefaultVoices   map[string]string            // lang -> voiceID
    Pronunciations  map[string]map[string]string // term -> lang -> replacement
    Slides          []Slide
}
```

### Slide

```go
type Slide struct {
    Title    string
    Notes    string
    Segments []Segment
}
```

### Segment

```go
type Segment struct {
    Text           map[string]string            // lang -> text
    Voice          map[string]string            // lang -> voiceID (override)
    PauseBefore    string                       // e.g., "500ms"
    PauseAfter     string
    Emphasis       string                       // "strong", "moderate", "reduced"
    Rate           string                       // "slow", "medium", "fast"
    Pitch          string                       // "low", "medium", "high"
    Pronunciations map[string]map[string]string // segment-level overrides
}
```

### CompiledSegment

```go
type CompiledSegment struct {
    SlideIndex    int
    SegmentIndex  int
    SlideTitle    string
    Text          string  // With pronunciations applied
    OriginalText  string
    VoiceID       string
    Language      string
    PauseBeforeMs int
    PauseAfterMs  int
    Emphasis      string
    Rate          string
    Pitch         string
}
```

## Functions

### Loading Scripts

```go
// Load from file
script, err := ttsscript.LoadScript("script.json")

// Parse from bytes
script, err := ttsscript.ParseScript(jsonData)

// Save to file
err := script.Save("output.json")
```

### Script Methods

```go
// Get all languages used
langs := script.Languages() // []string{"en", "es", "fr"}

// Count slides and segments
slideCount := script.SlideCount()
segmentCount := script.SegmentCount()

// Validate the script
issues := script.Validate() // []string of issues
```

### Compiler

```go
// Create compiler
compiler := ttsscript.NewCompiler()

// Configure defaults
compiler.DefaultPauseAfterSlide = "800ms"
compiler.DefaultPauseAfterSegment = "200ms"

// Add pronunciations
compiler.AddPronunciation("API", "en", "A P I")
compiler.AddPronunciations("en", map[string]string{
    "SDK": "S D K",
    "CLI": "C L I",
})

// Compile for a language
segments, err := compiler.Compile(script, "en")
```

### SSML Formatter

```go
formatter := ttsscript.NewSSMLFormatter()
formatter.Version = "1.1"
formatter.IncludeComments = true
formatter.IndentSpaces = 2

// Format compiled segments
ssml := formatter.Format(segments, "en")

// Or format directly from script
ssml, err := formatter.FormatScript(script, "en")
```

### ElevenLabs Formatter

```go
formatter := ttsscript.NewElevenLabsFormatter()
formatter.UsePauseMarkers = false

// Format compiled segments
jobs := formatter.Format(segments)

// Group by voice for batch processing
groups := formatter.GroupByVoice(jobs)

// Combine for single request (loses voice control)
text := formatter.CombineForSingleRequest(jobs)
```

### Batch Processing

```go
// Create batch config
config := ttsscript.NewBatchConfig("./output")
config.FilePrefix = "course"
config.IncludeLanguageInFilename = true

// Generate filenames
filename := config.GenerateFilename(job, "en")
// "./output/course_slide01_seg01_en.mp3"

// Generate manifest
manifest := ttsscript.GenerateManifest(jobs, config, "en")
```

### Utility Functions

```go
// Parse duration string to milliseconds
ms := ttsscript.ParseDuration("500ms") // 500
ms := ttsscript.ParseDuration("1.5s")  // 1500

// Format milliseconds to string
s := ttsscript.FormatDuration(500)  // "500ms"
s := ttsscript.FormatDuration(2000) // "2s"

// Group segments
byVoice := ttsscript.GroupByVoice(segments)
bySlide := ttsscript.GroupBySlide(segments)

// Combine text with pause markers
text := ttsscript.CombineText(segments)
```

### SSML Helpers

```go
// Generate SSML elements
break := ttsscript.SSMLBreak("500ms")
// <break time="500ms"/>

prosody := ttsscript.SSMLProsody("text", "slow", "+10%", "")
// <prosody rate="slow" pitch="+10%">text</prosody>

emphasis := ttsscript.SSMLEmphasis("important", "strong")
// <emphasis level="strong">important</emphasis>

sayAs := ttsscript.SSMLSayAs("123-456-7890", "telephone", "")
// <say-as interpret-as="telephone">123-456-7890</say-as>

phoneme := ttsscript.SSMLPhoneme("tomato", "ipa", "təˈmeɪtoʊ")
// <phoneme alphabet="ipa" ph="təˈmeɪtoʊ">tomato</phoneme>

sub := ttsscript.SSMLSub("API", "A P I")
// <sub alias="A P I">API</sub>

// Escape special characters
escaped := ttsscript.EscapeSSML("Tom & Jerry")
// "Tom &amp; Jerry"
```

## Example: Complete Workflow

```go
package main

import (
    "context"
    "fmt"
    "io"
    "os"

    elevenlabs "github.com/agentplexus/go-elevenlabs"
    "github.com/agentplexus/go-elevenlabs/ttsscript"
)

func main() {
    // Load script
    script, _ := ttsscript.LoadScript("course.json")

    // Create compiler with pronunciations
    compiler := ttsscript.NewCompiler()
    compiler.AddPronunciations("en", map[string]string{
        "API": "A P I",
        "SDK": "S D K",
    })

    // Compile for English
    segments, _ := compiler.Compile(script, "en")

    // Format for ElevenLabs
    formatter := ttsscript.NewElevenLabsFormatter()
    jobs := formatter.Format(segments)

    // Generate audio
    client, _ := elevenlabs.NewClient()
    ctx := context.Background()

    os.MkdirAll("./output", 0755)

    for i, job := range jobs {
        audio, _ := client.TextToSpeech().Simple(ctx, job.VoiceID, job.Text)

        f, _ := os.Create(fmt.Sprintf("./output/segment_%02d.mp3", i+1))
        io.Copy(f, audio)
        f.Close()

        fmt.Printf("Generated: segment_%02d.mp3\n", i+1)
    }
}
```

## See Also

- [TTS Script Authoring Guide](../guides/ttsscript.md) - Detailed authoring guide
- [LMS/Udemy Courses](../guides/lms-courses.md) - Using ttsscript for courses
