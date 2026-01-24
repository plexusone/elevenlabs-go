# Pronunciation Dictionaries

Ensure correct pronunciation of technical terms, names, and domain-specific vocabulary.

## Why Use Pronunciation Dictionaries?

Without pronunciation rules, TTS may mispronounce:

- **Acronyms**: "API" as "appy" instead of "A P I"
- **Technical terms**: "kubectl" as "kub-cuttle"
- **Brand names**: "nginx" incorrectly
- **Domain jargon**: Industry-specific terms

## Creating Dictionaries

### From a Map (Simplest)

```go
dict, err := client.Pronunciation().CreateFromMap(ctx, "Tech Terms", map[string]string{
    "ADK":     "Agent Development Kit",
    "API":     "A P I",
    "kubectl": "kube control",
    "nginx":   "engine X",
    "SQL":     "sequel",
})
```

### From a JSON File

Create a JSON file (`terms.json`):

```json
[
  {"grapheme": "ADK", "alias": "Agent Development Kit"},
  {"grapheme": "API", "alias": "A P I"},
  {"grapheme": "kubectl", "alias": "kube control"},
  {"grapheme": "nginx", "phoneme": "ˈɛndʒɪnˈɛks"}
]
```

Load and create:

```go
dict, err := client.Pronunciation().CreateFromJSON(ctx, "Tech Terms", "terms.json")
```

### With Full Options

```go
rules := elevenlabs.PronunciationRules{
    {Grapheme: "ADK", Alias: "Agent Development Kit"},
    {Grapheme: "nginx", Phoneme: "ˈɛndʒɪnˈɛks"},
}

dict, err := client.Pronunciation().Create(ctx, &elevenlabs.CreatePronunciationDictionaryRequest{
    Name:        "Tech Terms",
    Description: "Technical vocabulary for developer courses",
    Rules:       rules,
    Language:    "en-US",
})
```

## Rule Types

### Alias (Text Substitution)

The simpler option - specify replacement text:

```go
{Grapheme: "API", Alias: "A P I"}
// "API" will be read as "A P I"
```

### Phoneme (IPA)

For precise phonetic control using International Phonetic Alphabet:

```go
{Grapheme: "nginx", Phoneme: "ˈɛndʒɪnˈɛks"}
```

## Managing Dictionaries

### List All Dictionaries

```go
resp, err := client.Pronunciation().List(ctx, nil)
for _, dict := range resp.Dictionaries {
    fmt.Printf("%s: %s (%d rules)\n", dict.ID, dict.Name, dict.RulesCount)
}
```

### Get a Dictionary

```go
dict, err := client.Pronunciation().Get(ctx, dictionaryID)
```

### Rename a Dictionary

```go
err := client.Pronunciation().Rename(ctx, dictionaryID, "New Name")
```

### Remove Rules

```go
err := client.Pronunciation().RemoveRules(ctx, dictionaryID, []string{"API", "SQL"})
```

### Archive a Dictionary

```go
err := client.Pronunciation().Archive(ctx, dictionaryID)
```

### Download PLS File

Download the PLS (Pronunciation Lexicon Specification) XML file for a dictionary:

```go
// Download latest version
pls, err := client.Pronunciation().DownloadLatestPLS(ctx, dictionaryID)
if err != nil {
    log.Fatal(err)
}

// Save to file
f, _ := os.Create("dictionary.pls")
io.Copy(f, pls)

// Or download a specific version
pls, err := client.Pronunciation().GetVersionPLS(ctx, dictionaryID, versionID)
```

## Working with Rules Locally

### Load from JSON

```go
rules, err := elevenlabs.LoadRulesFromJSON("terms.json")
```

### Parse from JSON String

```go
jsonData := `[{"grapheme": "API", "alias": "A P I"}]`
rules, err := elevenlabs.ParseRulesFromJSON([]byte(jsonData))
```

### Create from Map

```go
rules := elevenlabs.RulesFromMap(map[string]string{
    "API": "A P I",
    "SQL": "sequel",
})
```

### Generate PLS XML

```go
rules := elevenlabs.PronunciationRules{
    {Grapheme: "API", Alias: "A P I"},
}

// Get as string
plsXML, err := rules.ToPLSString("en-US")

// Save to file
err = rules.SavePLS("terms.pls", "en-US")
```

### View Rules

```go
fmt.Println(rules.String())
// Output:
// API → A P I
// nginx → [ˈɛndʒɪnˈɛks]

// Get all graphemes
terms := rules.Graphemes()  // ["API", "nginx"]
```

## JSON File Format

```json
[
  {
    "grapheme": "ADK",
    "alias": "Agent Development Kit"
  },
  {
    "grapheme": "API",
    "alias": "A P I"
  },
  {
    "grapheme": "nginx",
    "phoneme": "ˈɛndʒɪnˈɛks"
  }
]
```

## Best Practices

1. **Use aliases for simplicity** - Phonemes only when needed
2. **Test pronunciations** - Generate sample audio to verify
3. **Organize by domain** - Separate dictionaries for different topics
4. **Version control your JSON** - Track changes to pronunciation rules
5. **Document unusual terms** - Add comments explaining why terms need rules
