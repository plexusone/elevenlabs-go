# History

Access and manage your generated audio history.

## List History

```go
resp, err := client.History().List(ctx, nil)
if err != nil {
    log.Fatal(err)
}

for _, item := range resp.Items {
    fmt.Printf("%s: %s (%s)\n", item.HistoryItemID, item.Text[:50], item.VoiceID)
}
```

### With Pagination

```go
resp, err := client.History().List(ctx, &elevenlabs.HistoryListOptions{
    PageSize: 20,
})

// Get next page
if resp.HasMore {
    nextResp, _ := client.History().List(ctx, &elevenlabs.HistoryListOptions{
        PageSize:             20,
        StartAfterHistoryID: resp.LastHistoryItemID,
    })
}
```

### Filter by Voice

```go
resp, err := client.History().List(ctx, &elevenlabs.HistoryListOptions{
    VoiceID: "21m00Tcm4TlvDq8ikWAM",
})
```

## History Item Object

| Field | Description |
|-------|-------------|
| `HistoryItemID` | Unique identifier |
| `VoiceID` | Voice used |
| `VoiceName` | Voice name |
| `Text` | Input text |
| `ModelID` | Model used |
| `DateUnix` | Creation timestamp |
| `CharacterCount` | Characters used |
| `ContentType` | MIME type |
| `State` | Processing state |

## Get a Specific Item

```go
item, err := client.History().Get(ctx, historyItemID)
```

## Download Audio

```go
audio, err := client.History().GetAudio(ctx, historyItemID)
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("downloaded.mp3")
defer f.Close()
io.Copy(f, audio)
```

## Delete History Item

```go
err := client.History().Delete(ctx, historyItemID)
```

## Use Cases

### Re-download Lost Audio

```go
// Find item by text content
resp, _ := client.History().List(ctx, nil)
for _, item := range resp.Items {
    if strings.Contains(item.Text, "specific phrase") {
        audio, _ := client.History().GetAudio(ctx, item.HistoryItemID)
        // Save audio
    }
}
```

### Track Usage Over Time

```go
resp, _ := client.History().List(ctx, &elevenlabs.HistoryListOptions{
    PageSize: 100,
})

var totalChars int
for _, item := range resp.Items {
    totalChars += item.CharacterCount
}
fmt.Printf("Total characters used: %d\n", totalChars)
```

### Clean Up Old Items

```go
cutoff := time.Now().AddDate(0, -1, 0)  // 1 month ago

resp, _ := client.History().List(ctx, nil)
for _, item := range resp.Items {
    if time.Unix(item.DateUnix, 0).Before(cutoff) {
        client.History().Delete(ctx, item.HistoryItemID)
    }
}
```
