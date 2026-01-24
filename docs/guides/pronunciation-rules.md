# Pronunciation Rules Guide

A complete guide to managing pronunciation dictionaries for correct term pronunciation.

## Why Pronunciation Rules?

Text-to-Speech engines may mispronounce:

| Term | Without Rules | With Rules |
|------|---------------|------------|
| API | "appy" | "A P I" |
| SQL | "squeal" | "sequel" |
| nginx | "ningks" | "engine X" |
| kubectl | "kub-cuttle" | "kube control" |
| OAuth | "oh-ath" | "oh auth" |

## JSON-Based Workflow

### 1. Create a Rules File

`pronunciation-rules.json`:

```json
[
  {"grapheme": "API", "alias": "A P I"},
  {"grapheme": "SDK", "alias": "S D K"},
  {"grapheme": "CLI", "alias": "C L I"},
  {"grapheme": "GUI", "alias": "gooey"},
  {"grapheme": "SQL", "alias": "sequel"},
  {"grapheme": "OAuth", "alias": "oh auth"},
  {"grapheme": "JSON", "alias": "jay son"},
  {"grapheme": "YAML", "alias": "yammel"},
  {"grapheme": "nginx", "alias": "engine X"},
  {"grapheme": "kubectl", "alias": "kube control"},
  {"grapheme": "etcd", "alias": "et see dee"},
  {"grapheme": "gRPC", "alias": "gee R P C"}
]
```

### 2. Load and Create Dictionary

```go
dict, err := client.Pronunciation().CreateFromJSON(ctx,
    "Tech Terms",
    "pronunciation-rules.json")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created dictionary %s with %d rules\n", dict.ID, dict.RulesCount)
```

## Rule Types

### Alias (Recommended)

Simple text substitution - the easiest approach:

```json
{"grapheme": "API", "alias": "A P I"}
```

The text "API" will be replaced with "A P I" before TTS processing.

**Tips for aliases:**

- Space out letters: `"API"` → `"A P I"`
- Use phonetic spelling: `"nginx"` → `"engine X"`
- Break compound words: `"OAuth"` → `"oh auth"`

### Phoneme (Advanced)

Use IPA (International Phonetic Alphabet) for precise control:

```json
{"grapheme": "nginx", "phoneme": "ˈɛndʒɪnˈɛks"}
```

**When to use phonemes:**

- When aliases don't produce correct pronunciation
- For names with unusual pronunciation
- For non-English words

## Working with Rules in Go

### Create from Map

```go
rules := elevenlabs.RulesFromMap(map[string]string{
    "API":     "A P I",
    "kubectl": "kube control",
})

dict, err := client.Pronunciation().Create(ctx, &elevenlabs.CreatePronunciationDictionaryRequest{
    Name:  "Quick Terms",
    Rules: rules,
})
```

### Load from JSON File

```go
rules, err := elevenlabs.LoadRulesFromJSON("terms.json")
if err != nil {
    log.Fatal(err)
}

// View rules
fmt.Println(rules.String())
// Output:
// API → A P I
// kubectl → kube control
```

### Parse from JSON String

```go
jsonData := `[
    {"grapheme": "API", "alias": "A P I"},
    {"grapheme": "SDK", "alias": "S D K"}
]`

rules, err := elevenlabs.ParseRulesFromJSON([]byte(jsonData))
```

### Generate PLS XML

```go
rules := elevenlabs.PronunciationRules{
    {Grapheme: "API", Alias: "A P I"},
}

// Get XML string
xml, err := rules.ToPLSString("en-US")
fmt.Println(xml)

// Save to file
err = rules.SavePLS("terms.pls", "en-US")
```

Generated PLS:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<lexicon version="1.0"
         xmlns="http://www.w3.org/2005/01/pronunciation-lexicon"
         alphabet="ipa" xml:lang="en-US">
  <lexeme>
    <grapheme>API</grapheme>
    <alias>A P I</alias>
  </lexeme>
</lexicon>
```

## Managing Dictionaries

### List All Dictionaries

```go
resp, err := client.Pronunciation().List(ctx, nil)
for _, d := range resp.Dictionaries {
    fmt.Printf("%s: %s (%d rules)\n", d.ID, d.Name, d.RulesCount)
}
```

### Update a Dictionary

To add new rules, create a new dictionary version (append to your JSON, re-upload).

### Remove Specific Rules

```go
err := client.Pronunciation().RemoveRules(ctx, dictID, []string{"API", "SDK"})
```

### Rename Dictionary

```go
err := client.Pronunciation().Rename(ctx, dictID, "New Name")
```

### Archive Dictionary

```go
err := client.Pronunciation().Archive(ctx, dictID)
```

### Download PLS File

Export the dictionary as a PLS (Pronunciation Lexicon Specification) XML file:

```go
// Download latest version
pls, err := client.Pronunciation().DownloadLatestPLS(ctx, dictID)
if err != nil {
    log.Fatal(err)
}

// Save to file
f, _ := os.Create("dictionary.pls")
io.Copy(f, pls)
f.Close()

// Or download a specific version
pls, err := client.Pronunciation().GetVersionPLS(ctx, dictID, versionID)
```

## Domain-Specific Examples

### Software Development

```json
[
  {"grapheme": "API", "alias": "A P I"},
  {"grapheme": "REST", "alias": "rest"},
  {"grapheme": "GraphQL", "alias": "graph Q L"},
  {"grapheme": "npm", "alias": "N P M"},
  {"grapheme": "pip", "alias": "pip"},
  {"grapheme": "git", "alias": "git"},
  {"grapheme": "GitHub", "alias": "git hub"},
  {"grapheme": "VS Code", "alias": "V S code"}
]
```

### Cloud/DevOps

```json
[
  {"grapheme": "AWS", "alias": "A W S"},
  {"grapheme": "GCP", "alias": "G C P"},
  {"grapheme": "Azure", "alias": "azher"},
  {"grapheme": "K8s", "alias": "kubernetes"},
  {"grapheme": "CI/CD", "alias": "C I C D"},
  {"grapheme": "DevOps", "alias": "dev ops"},
  {"grapheme": "IaC", "alias": "I A C"}
]
```

### Data Science

```json
[
  {"grapheme": "ML", "alias": "M L"},
  {"grapheme": "AI", "alias": "A I"},
  {"grapheme": "NLP", "alias": "N L P"},
  {"grapheme": "GPU", "alias": "G P U"},
  {"grapheme": "TPU", "alias": "T P U"},
  {"grapheme": "PyTorch", "alias": "pie torch"},
  {"grapheme": "TensorFlow", "alias": "tensor flow"},
  {"grapheme": "NumPy", "alias": "num pie"}
]
```

## Best Practices

1. **Version control your JSON** - Track changes over time
2. **Test pronunciations** - Generate sample audio to verify
3. **Document unusual rules** - Add comments explaining why
4. **Organize by domain** - Separate dictionaries for different topics
5. **Use aliases first** - Only use phonemes when necessary
6. **Consider context** - Same acronym may have different pronunciations
