# Projects (Studio)

Create long-form audio content organized into chapters - ideal for audiobooks, courses, and podcasts.

## Overview

Projects allow you to:

- Organize content into chapters
- Apply consistent voice settings across content
- Convert chapters individually or all at once
- Download completed audio as snapshots

## Creating a Project

### Basic Project

```go
project, err := client.Projects().Create(ctx, &elevenlabs.CreateProjectRequest{
    Name:        "My Course",
    Description: "Introduction to Go Programming",
    Language:    "en",
})
```

### With Voice Settings

```go
project, err := client.Projects().Create(ctx, &elevenlabs.CreateProjectRequest{
    Name:                    "My Course",
    DefaultModelID:          "eleven_multilingual_v2",
    DefaultParagraphVoiceID: "21m00Tcm4TlvDq8ikWAM",
    DefaultTitleVoiceID:     "21m00Tcm4TlvDq8ikWAM",
    QualityPreset:           "high",
    AutoConvert:             false,
})
```

### From URL

```go
project, err := client.Projects().Create(ctx, &elevenlabs.CreateProjectRequest{
    Name:    "My Article",
    FromURL: "https://example.com/article",
})
```

## Listing Projects

```go
projects, err := client.Projects().List(ctx)
for _, p := range projects {
    fmt.Printf("%s: %s\n", p.ProjectID, p.Name)
}
```

## Project Object

| Field | Description |
|-------|-------------|
| `ProjectID` | Unique identifier |
| `Name` | Project name |
| `Description` | Project description |
| `Author` | Author name |
| `Language` | Two-letter language code |
| `DefaultModelID` | Default TTS model |
| `DefaultParagraphVoiceID` | Default voice for paragraphs |
| `DefaultTitleVoiceID` | Default voice for titles |
| `CreatedAt` | Creation timestamp |
| `AccessLevel` | Access permissions |

## Working with Chapters

### List Chapters

```go
chapters, err := client.Projects().ListChapters(ctx, projectID)
for _, ch := range chapters {
    fmt.Printf("%s: %s (state: %s)\n", ch.ChapterID, ch.Name, ch.State)
}
```

### Convert a Chapter

```go
err := client.Projects().ConvertChapter(ctx, projectID, chapterID)
```

### Delete a Chapter

```go
err := client.Projects().DeleteChapter(ctx, projectID, chapterID)
```

## Chapter Object

| Field | Description |
|-------|-------------|
| `ChapterID` | Unique identifier |
| `Name` | Chapter name |
| `State` | Current state |
| `ConversionProgress` | Progress percentage (0-100) |
| `LastConversionError` | Error message if failed |

## Converting Projects

### Convert Entire Project

```go
err := client.Projects().Convert(ctx, projectID)
```

### Check Conversion Status

```go
chapters, _ := client.Projects().ListChapters(ctx, projectID)
for _, ch := range chapters {
    fmt.Printf("%s: %.0f%% complete\n", ch.Name, ch.ConversionProgress)
}
```

## Snapshots

Snapshots are frozen versions of converted audio.

### List Project Snapshots

```go
snapshots, err := client.Projects().ListSnapshots(ctx, projectID)
for _, snap := range snapshots {
    fmt.Printf("%s: %s (created: %s)\n",
        snap.ProjectSnapshotID, snap.Name, snap.CreatedAt)
}
```

### Download Snapshot Archive

```go
reader, err := client.Projects().DownloadSnapshotArchive(ctx, projectID, snapshotID)
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("project.zip")
defer f.Close()
io.Copy(f, reader)
```

### List Chapter Snapshots

```go
snapshots, err := client.Projects().ListChapterSnapshots(ctx, projectID, chapterID)
```

### Stream Chapter Audio

```go
audio, err := client.Projects().StreamChapterAudio(ctx, projectID, chapterID, snapshotID)
```

## Updating Projects

```go
err := client.Projects().Update(ctx, projectID, &elevenlabs.UpdateProjectRequest{
    Name:                    "Updated Name",
    DefaultParagraphVoiceID: "newVoiceID",
    DefaultTitleVoiceID:     "newVoiceID",
})
```

## Deleting Projects

```go
err := client.Projects().Delete(ctx, projectID)
```

## Quality Presets

| Preset | Description |
|--------|-------------|
| `standard` | 128kbps, 44.1kHz |
| `high` | 192kbps, 44.1kHz |
| `ultra` | 192kbps, enhanced |
| `ultra lossless` | 705.6kbps, lossless |

## Workflow Example

```go
// 1. Create project
project, _ := client.Projects().Create(ctx, &elevenlabs.CreateProjectRequest{
    Name:     "Go Programming Course",
    Language: "en",
})

// 2. List chapters (added via web UI or API)
chapters, _ := client.Projects().ListChapters(ctx, project.ProjectID)

// 3. Convert all chapters
for _, ch := range chapters {
    client.Projects().ConvertChapter(ctx, project.ProjectID, ch.ChapterID)
}

// 4. Wait for conversion (poll status)
// ...

// 5. Download completed project
snapshots, _ := client.Projects().ListSnapshots(ctx, project.ProjectID)
if len(snapshots) > 0 {
    reader, _ := client.Projects().DownloadSnapshotArchive(ctx,
        project.ProjectID, snapshots[0].ProjectSnapshotID)

    f, _ := os.Create("course.zip")
    io.Copy(f, reader)
    f.Close()
}
```
