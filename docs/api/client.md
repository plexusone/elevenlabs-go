# Client API Reference

## Client

The main client for interacting with the ElevenLabs API.

### Constructor

```go
func NewClient(opts ...Option) (*Client, error)
```

Creates a new client with optional configuration.

**Options:**

| Option | Description |
|--------|-------------|
| `WithAPIKey(key string)` | Set API key |
| `WithBaseURL(url string)` | Set base URL |
| `WithHTTPClient(client *http.Client)` | Set HTTP client |
| `WithTimeout(timeout time.Duration)` | Set request timeout |

**Example:**

```go
client, err := elevenlabs.NewClient(
    elevenlabs.WithAPIKey("your-api-key"),
    elevenlabs.WithTimeout(5 * time.Minute),
)
```

### Service Accessors

| Method | Returns | Description |
|--------|---------|-------------|
| `TextToSpeech()` | `*TextToSpeechService` | Text-to-speech operations |
| `Voices()` | `*VoicesService` | Voice management |
| `Models()` | `*ModelsService` | Model listing |
| `History()` | `*HistoryService` | Generation history |
| `User()` | `*UserService` | User/subscription info |
| `Dubbing()` | `*DubbingService` | Video dubbing |
| `SoundEffects()` | `*SoundEffectsService` | Sound effect generation |
| `Pronunciation()` | `*PronunciationService` | Pronunciation dictionaries |
| `Projects()` | `*ProjectsService` | Studio projects |
| `API()` | `*api.Client` | Raw ogen client |

### Constants

```go
const Version = "0.1.0"
const DefaultBaseURL = "https://api.elevenlabs.io"
const DefaultModelID = "eleven_multilingual_v2"
```

## TextToSpeechService

### Generate

```go
func (s *TextToSpeechService) Generate(ctx context.Context, req *TTSRequest) (*TTSResponse, error)
```

Generate speech with full control over options.

### Simple

```go
func (s *TextToSpeechService) Simple(ctx context.Context, voiceID, text string) (io.Reader, error)
```

Generate speech with default settings.

## VoicesService

### List

```go
func (s *VoicesService) List(ctx context.Context) ([]*Voice, error)
```

### Get

```go
func (s *VoicesService) Get(ctx context.Context, voiceID string) (*Voice, error)
```

### GetSettings

```go
func (s *VoicesService) GetSettings(ctx context.Context, voiceID string) (*VoiceSettings, error)
```

### GetDefaultSettings

```go
func (s *VoicesService) GetDefaultSettings(ctx context.Context) (*VoiceSettings, error)
```

## SoundEffectsService

### Generate

```go
func (s *SoundEffectsService) Generate(ctx context.Context, req *SoundEffectRequest) (*SoundEffectResponse, error)
```

### Simple

```go
func (s *SoundEffectsService) Simple(ctx context.Context, description string) (io.Reader, error)
```

### GenerateLoop

```go
func (s *SoundEffectsService) GenerateLoop(ctx context.Context, description string, durationSeconds float64) (io.Reader, error)
```

## PronunciationService

### List

```go
func (s *PronunciationService) List(ctx context.Context, opts *PronunciationDictionaryListOptions) (*PronunciationDictionaryListResponse, error)
```

### Get

```go
func (s *PronunciationService) Get(ctx context.Context, dictionaryID string) (*PronunciationDictionary, error)
```

### Create

```go
func (s *PronunciationService) Create(ctx context.Context, req *CreatePronunciationDictionaryRequest) (*PronunciationDictionary, error)
```

### CreateFromJSON

```go
func (s *PronunciationService) CreateFromJSON(ctx context.Context, name, jsonFilePath string) (*PronunciationDictionary, error)
```

### CreateFromMap

```go
func (s *PronunciationService) CreateFromMap(ctx context.Context, name string, rules map[string]string) (*PronunciationDictionary, error)
```

### RemoveRules

```go
func (s *PronunciationService) RemoveRules(ctx context.Context, dictionaryID string, ruleStrings []string) error
```

### Rename

```go
func (s *PronunciationService) Rename(ctx context.Context, dictionaryID, newName string) error
```

### Archive

```go
func (s *PronunciationService) Archive(ctx context.Context, dictionaryID string) error
```

## ProjectsService

### List

```go
func (s *ProjectsService) List(ctx context.Context) ([]*Project, error)
```

### Create

```go
func (s *ProjectsService) Create(ctx context.Context, req *CreateProjectRequest) (*Project, error)
```

### Update

```go
func (s *ProjectsService) Update(ctx context.Context, projectID string, req *UpdateProjectRequest) error
```

### Delete

```go
func (s *ProjectsService) Delete(ctx context.Context, projectID string) error
```

### Convert

```go
func (s *ProjectsService) Convert(ctx context.Context, projectID string) error
```

### ListChapters

```go
func (s *ProjectsService) ListChapters(ctx context.Context, projectID string) ([]*Chapter, error)
```

### ConvertChapter

```go
func (s *ProjectsService) ConvertChapter(ctx context.Context, projectID, chapterID string) error
```

### DeleteChapter

```go
func (s *ProjectsService) DeleteChapter(ctx context.Context, projectID, chapterID string) error
```

### ListSnapshots

```go
func (s *ProjectsService) ListSnapshots(ctx context.Context, projectID string) ([]*ProjectSnapshot, error)
```

### DownloadSnapshotArchive

```go
func (s *ProjectsService) DownloadSnapshotArchive(ctx context.Context, projectID, snapshotID string) (io.Reader, error)
```

## Helper Functions

### PronunciationRules

```go
func LoadRulesFromJSON(filename string) (PronunciationRules, error)
func ParseRulesFromJSON(data []byte) (PronunciationRules, error)
func RulesFromMap(m map[string]string) PronunciationRules

func (rules PronunciationRules) ToPLS(language string) ([]byte, error)
func (rules PronunciationRules) ToPLSString(language string) (string, error)
func (rules PronunciationRules) SavePLS(filename, language string) error
func (rules PronunciationRules) Graphemes() []string
func (rules PronunciationRules) String() string
```

### VoiceSettings

```go
func DefaultVoiceSettings() *VoiceSettings
```

Returns default voice settings (Stability: 0.5, SimilarityBoost: 0.75).
