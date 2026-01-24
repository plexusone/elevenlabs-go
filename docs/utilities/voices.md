# Voice Reference

The `voices` package provides reference information for ElevenLabs pre-made voices.

## Installation

```go
import "github.com/agentplexus/go-elevenlabs/voices"
```

## Quick Start

### Use Voice Constants

```go
import "github.com/agentplexus/go-elevenlabs/voices"

// Use constants instead of hard-coded IDs
audio, err := client.TextToSpeech().Simple(ctx, voices.Rachel, "Hello world")

// Other popular voices
audio, err := client.TextToSpeech().Simple(ctx, voices.Adam, "Deep male voice")
audio, err := client.TextToSpeech().Simple(ctx, voices.George, "British accent")
```

### Look Up Voice Metadata

```go
// Get voice by ID
v := voices.GetVoice(voices.Rachel)
fmt.Printf("Name: %s\n", v.Name)           // "Rachel"
fmt.Printf("Gender: %s\n", v.Gender)       // "female"
fmt.Printf("Accent: %s\n", v.Accent)       // "American"
fmt.Printf("Use Case: %s\n", v.UseCase)    // "Narration, audiobooks"

// Get voice by name (case-insensitive)
v := voices.GetVoiceByName("rachel")
```

### Filter Voices

```go
// Find female voices
females := voices.FilterByGender("female")

// Find British voices
british := voices.FilterByAccent("British")

// Find young voices
young := voices.FilterByAge("young")
```

## Pre-made Voice Constants

### Female Voices

| Constant | Name | Description | Accent | Age |
|----------|------|-------------|--------|-----|
| `Rachel` | Rachel | Calm and composed | American | Young |
| `Domi` | Domi | Strong and confident | American | Young |
| `Bella` | Bella | Soft and warm | American | Young |
| `Elli` | Elli | Emotional and expressive | American | Young |
| `Nicole` | Nicole | Soft and whispery | American | Young |
| `Emily` | Emily | Calm and professional | American | Young |
| `Freya` | Freya | Expressive and clear | American | Young |
| `Gigi` | Gigi | Childlike and playful | American | Young |
| `Grace` | Grace | Southern and sweet | American Southern | Young |
| `Dorothy` | Dorothy | Pleasant and refined | British | Young |
| `Charlotte` | Charlotte | Seductive and sophisticated | Swedish | Middle-aged |
| `Matilda` | Matilda | Warm and friendly | American | Middle-aged |
| `Lily` | Lily | Raspy British | British | Middle-aged |
| `Serena` | Serena | Pleasant and calm | American | Middle-aged |
| `Glinda` | Glinda | Theatrical witch-like | American | Middle-aged |

### Male Voices

| Constant | Name | Description | Accent | Age |
|----------|------|-------------|--------|-----|
| `Antoni` | Antoni | Well-rounded and professional | American | Young |
| `Josh` | Josh | Deep and authoritative | American | Young |
| `Sam` | Sam | Raspy and casual | American | Young |
| `Ethan` | Ethan | Energetic and youthful | American | Young |
| `Jeremy` | Jeremy | Conversational and natural | American | Young |
| `Harry` | Harry | Anxious energy | American | Young |
| `Liam` | Liam | Articulate and clear | American | Young |
| `Dave` | Dave | Conversational British-Essex | British | Young |
| `Arnold` | Arnold | Crisp and confident | American | Middle-aged |
| `Adam` | Adam | Deep and warm | American | Middle-aged |
| `Brian` | Brian | Deep narrator quality | American | Middle-aged |
| `Drew` | Drew | Well-rounded and versatile | American | Middle-aged |
| `Paul` | Paul | Professional reporter style | American | Middle-aged |
| `Chris` | Chris | Casual and relaxed | American | Middle-aged |
| `Clyde` | Clyde | Gruff war veteran | American | Middle-aged |
| `Callum` | Callum | Intense and dramatic | Transatlantic | Middle-aged |
| `George` | George | Warm and refined | British | Middle-aged |
| `Joseph` | Joseph | Authoritative British | British | Middle-aged |
| `Michael` | Michael | Wise and grandfatherly | American | Old |
| `Jessie` | Jessie | Raspy and weathered | American | Old |
| `Fin` | Fin | Weathered Irish sailor | Irish | Old |
| `James` | James | Warm Australian | Australian | Old |

### Non-Binary Voices

| Constant | Name | Description | Accent | Age |
|----------|------|-------------|--------|-----|
| `River` | River | Modern and inclusive | American | Young |

## Voice Type

```go
type Voice struct {
    ID          string // Unique voice identifier
    Name        string // Display name
    Description string // Voice characteristics
    Gender      string // male, female, non-binary
    Age         string // young, middle-aged, old
    Accent      string // American, British, etc.
    UseCase     string // Suggested use cases
    Category    string // premade, cloned, designed
}
```

## Functions

### Get All Voices

```go
allVoices := voices.PremadeVoices()
for _, v := range allVoices {
    fmt.Printf("%s (%s): %s\n", v.Name, v.ID, v.Description)
}
```

### Look Up by ID

```go
v := voices.GetVoice("21m00Tcm4TlvDq8ikWAM")
if v != nil {
    fmt.Printf("Found: %s\n", v.Name)
}
```

### Look Up by Name

```go
v := voices.GetVoiceByName("Rachel") // Case-insensitive
v := voices.GetVoiceByName("rachel") // Also works
```

### Filter by Attribute

```go
// By gender
females := voices.FilterByGender("female")
males := voices.FilterByGender("male")

// By accent (partial match)
british := voices.FilterByAccent("British")
american := voices.FilterByAccent("American")

// By age
young := voices.FilterByAge("young")
middleAged := voices.FilterByAge("middle-aged")
old := voices.FilterByAge("old")
```

## Use Case Recommendations

### Narration & Audiobooks

- **Rachel** - Calm, young female
- **Adam** - Deep, warm male
- **Brian** - Deep narrator quality
- **George** - Refined British male

### Business & Education

- **Antoni** - Professional, warm
- **Matilda** - Friendly, approachable
- **Emily** - Professional female
- **Liam** - Articulate, clear

### Podcasts & Casual Content

- **Bella** - Warm and friendly
- **Jeremy** - Conversational
- **Sam** - Casual, raspy
- **Dave** - British conversational

### Character Voices & Gaming

- **Clyde** - Gruff war veteran
- **Glinda** - Theatrical witch
- **Fin** - Irish sailor
- **Harry** - Anxious character

### Documentaries & News

- **Josh** - Authoritative
- **Joseph** - British authority
- **Paul** - Reporter style
- **Callum** - Dramatic narration

## Integration with ttsscript

```go
import (
    "github.com/agentplexus/go-elevenlabs/ttsscript"
    "github.com/agentplexus/go-elevenlabs/voices"
)

script := &ttsscript.Script{
    DefaultVoices: map[string]string{
        "en": voices.Rachel,
        "es": voices.Bella,
    },
    // ...
}
```

## JSON Reference

The package includes a `voices.json` file with the same data for use in other tools:

```json
{
  "voices": [
    {
      "id": "21m00Tcm4TlvDq8ikWAM",
      "name": "Rachel",
      "description": "Calm and composed",
      "gender": "female",
      "age": "young",
      "accent": "American",
      "use_case": "Narration, audiobooks"
    }
  ]
}
```

## Note

Voice IDs and availability may change over time. For the authoritative list of voices available to your account, use:

```go
voices, err := client.Voices().List(ctx)
```

The constants in this package are based on commonly available pre-made voices and are provided for convenience.
