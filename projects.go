package elevenlabs

import (
	"context"
	"io"
	"time"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// ProjectsService handles Studio Projects operations.
// Projects (formerly known as "Studio") allow you to create long-form
// audio content like audiobooks, podcasts, and video course narration
// organized into chapters.
type ProjectsService struct {
	client *Client
}

// Project represents a Studio project.
type Project struct {
	// ProjectID is the unique identifier.
	ProjectID string

	// Name is the project name.
	Name string

	// Description is the project description.
	Description string

	// Author is the project author.
	Author string

	// Language is the two-letter language code (ISO 639-1).
	Language string

	// DefaultModelID is the default model for TTS.
	DefaultModelID string

	// DefaultParagraphVoiceID is the default voice for paragraphs.
	DefaultParagraphVoiceID string

	// DefaultTitleVoiceID is the default voice for titles.
	DefaultTitleVoiceID string

	// ContentType is the content type (e.g., "Novel", "Short Story").
	ContentType string

	// CoverImageURL is the cover image URL.
	CoverImageURL string

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time

	// CanBeDownloaded indicates if the project can be downloaded.
	CanBeDownloaded bool

	// AccessLevel is the access level of the project.
	AccessLevel string
}

// Chapter represents a chapter within a project.
type Chapter struct {
	// ChapterID is the unique identifier.
	ChapterID string

	// Name is the chapter name.
	Name string

	// ConversionProgress is the conversion progress percentage.
	ConversionProgress float64

	// State is the current state.
	State string

	// LastConversionError is the last conversion error if any.
	LastConversionError string
}

// ChapterSnapshot represents a snapshot of a chapter.
type ChapterSnapshot struct {
	// ChapterSnapshotID is the unique identifier.
	ChapterSnapshotID string

	// ProjectID is the parent project ID.
	ProjectID string

	// ChapterID is the chapter ID.
	ChapterID string

	// Name is the snapshot name.
	Name string

	// CreatedAt is when the snapshot was created.
	CreatedAt time.Time
}

// ProjectSnapshot represents a snapshot of a project.
type ProjectSnapshot struct {
	// ProjectSnapshotID is the unique identifier.
	ProjectSnapshotID string

	// ProjectID is the parent project ID.
	ProjectID string

	// Name is the snapshot name.
	Name string

	// CreatedAt is when the snapshot was created.
	CreatedAt time.Time
}

// CreateProjectRequest contains options for creating a project.
type CreateProjectRequest struct {
	// Name is the project name (required).
	Name string

	// Description is an optional description.
	Description string

	// Author is an optional author name.
	Author string

	// Language is the two-letter language code (ISO 639-1).
	Language string

	// DefaultModelID is the model to use for TTS.
	DefaultModelID string

	// DefaultParagraphVoiceID is the default voice for paragraphs.
	DefaultParagraphVoiceID string

	// DefaultTitleVoiceID is the default voice for titles.
	DefaultTitleVoiceID string

	// FromURL is a URL to extract content from.
	FromURL string

	// ContentType is the content type (e.g., "Novel", "Short Story").
	ContentType string

	// Genres is a list of genres.
	Genres []string

	// QualityPreset is the output quality: "standard", "high", "ultra", "ultra lossless".
	QualityPreset string

	// AutoConvert automatically converts the project to audio.
	AutoConvert bool
}

// Validate validates the create request.
func (r *CreateProjectRequest) Validate() error {
	if r.Name == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}
	return nil
}

// UpdateProjectRequest contains options for updating a project.
type UpdateProjectRequest struct {
	// Name is the new project name (required).
	Name string

	// DefaultParagraphVoiceID is the new default paragraph voice (required).
	DefaultParagraphVoiceID string

	// DefaultTitleVoiceID is the new default title voice (required).
	DefaultTitleVoiceID string

	// Author is an optional author name.
	Author string

	// Title is an optional title.
	Title string
}

// List returns all projects.
func (s *ProjectsService) List(ctx context.Context) ([]*Project, error) {
	resp, err := s.client.apiClient.GetProjects(ctx, api.GetProjectsParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.GetProjectsResponseModel:
		projects := make([]*Project, 0, len(r.Projects))
		for _, p := range r.Projects {
			proj := projectFromAPI(&p)
			projects = append(projects, proj)
		}
		return projects, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Create creates a new project.
func (s *ProjectsService) Create(ctx context.Context, req *CreateProjectRequest) (*Project, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	body := &api.BodyCreateStudioProjectV1StudioProjectsPostMultipart{
		Name: req.Name,
	}

	if req.Description != "" {
		body.Description = api.NewOptNilString(req.Description)
	}
	if req.Author != "" {
		body.Author = api.NewOptNilString(req.Author)
	}
	if req.Language != "" {
		body.Language = api.NewOptNilString(req.Language)
	}
	if req.DefaultModelID != "" {
		body.DefaultModelID = api.NewOptNilString(req.DefaultModelID)
	}
	if req.DefaultParagraphVoiceID != "" {
		body.DefaultParagraphVoiceID = api.NewOptNilString(req.DefaultParagraphVoiceID)
	}
	if req.DefaultTitleVoiceID != "" {
		body.DefaultTitleVoiceID = api.NewOptNilString(req.DefaultTitleVoiceID)
	}
	if req.FromURL != "" {
		body.FromURL = api.NewOptNilString(req.FromURL)
	}
	if req.ContentType != "" {
		body.ContentType = api.NewOptNilString(req.ContentType)
	}
	if len(req.Genres) > 0 {
		body.Genres = req.Genres
	}
	if req.QualityPreset != "" {
		body.QualityPreset = api.NewOptString(req.QualityPreset)
	}
	if req.AutoConvert {
		body.AutoConvert = api.NewOptBool(true)
	}

	resp, err := s.client.apiClient.AddProject(ctx, body, api.AddProjectParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.AddProjectResponseModel:
		return projectFromAPI(&r.Project), nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Update updates a project.
// Note: Name, DefaultParagraphVoiceID, and DefaultTitleVoiceID are required fields.
func (s *ProjectsService) Update(ctx context.Context, projectID string, req *UpdateProjectRequest) error {
	if projectID == "" {
		return &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}
	if req.DefaultParagraphVoiceID == "" {
		return &ValidationError{Field: "default_paragraph_voice_id", Message: "cannot be empty"}
	}
	if req.DefaultTitleVoiceID == "" {
		return &ValidationError{Field: "default_title_voice_id", Message: "cannot be empty"}
	}

	body := &api.BodyUpdateStudioProjectV1StudioProjectsProjectIDPost{
		Name:                    req.Name,
		DefaultParagraphVoiceID: req.DefaultParagraphVoiceID,
		DefaultTitleVoiceID:     req.DefaultTitleVoiceID,
	}

	if req.Author != "" {
		body.Author = api.NewOptNilString(req.Author)
	}
	if req.Title != "" {
		body.Title = api.NewOptNilString(req.Title)
	}

	_, err := s.client.apiClient.EditProject(ctx, body, api.EditProjectParams{
		ProjectID: projectID,
	})
	return err
}

// Delete deletes a project.
func (s *ProjectsService) Delete(ctx context.Context, projectID string) error {
	if projectID == "" {
		return &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}

	_, err := s.client.apiClient.DeleteProject(ctx, api.DeleteProjectParams{
		ProjectID: projectID,
	})
	return err
}

// Convert initiates conversion of a project to audio.
func (s *ProjectsService) Convert(ctx context.Context, projectID string) error {
	if projectID == "" {
		return &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}

	_, err := s.client.apiClient.ConvertProjectEndpoint(ctx, api.ConvertProjectEndpointParams{
		ProjectID: projectID,
	})
	return err
}

// ListChapters returns all chapters in a project.
func (s *ProjectsService) ListChapters(ctx context.Context, projectID string) ([]*Chapter, error) {
	if projectID == "" {
		return nil, &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetChapters(ctx, api.GetChaptersParams{
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.GetChaptersResponseModel:
		chapters := make([]*Chapter, 0, len(r.Chapters))
		for _, c := range r.Chapters {
			ch := &Chapter{
				ChapterID: c.ChapterID,
				Name:      c.Name,
				State:     string(c.State),
			}
			if c.ConversionProgress.Set && !c.ConversionProgress.Null {
				ch.ConversionProgress = c.ConversionProgress.Value
			}
			if c.LastConversionError.Set && !c.LastConversionError.Null {
				ch.LastConversionError = c.LastConversionError.Value
			}
			chapters = append(chapters, ch)
		}
		return chapters, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// ConvertChapter initiates conversion of a chapter to audio.
func (s *ProjectsService) ConvertChapter(ctx context.Context, projectID, chapterID string) error {
	if projectID == "" {
		return &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}
	if chapterID == "" {
		return &ValidationError{Field: "chapter_id", Message: "cannot be empty"}
	}

	_, err := s.client.apiClient.ConvertChapterEndpoint(ctx, api.ConvertChapterEndpointParams{
		ProjectID: projectID,
		ChapterID: chapterID,
	})
	return err
}

// DeleteChapter deletes a chapter from a project.
func (s *ProjectsService) DeleteChapter(ctx context.Context, projectID, chapterID string) error {
	if projectID == "" {
		return &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}
	if chapterID == "" {
		return &ValidationError{Field: "chapter_id", Message: "cannot be empty"}
	}

	_, err := s.client.apiClient.DeleteChapterEndpoint(ctx, api.DeleteChapterEndpointParams{
		ProjectID: projectID,
		ChapterID: chapterID,
	})
	return err
}

// ListSnapshots returns all snapshots for a project.
func (s *ProjectsService) ListSnapshots(ctx context.Context, projectID string) ([]*ProjectSnapshot, error) {
	if projectID == "" {
		return nil, &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetProjectSnapshots(ctx, api.GetProjectSnapshotsParams{
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.ProjectSnapshotsResponseModel:
		snapshots := make([]*ProjectSnapshot, 0, len(r.Snapshots))
		for _, snap := range r.Snapshots {
			snapshots = append(snapshots, &ProjectSnapshot{
				ProjectSnapshotID: snap.ProjectSnapshotID,
				ProjectID:         snap.ProjectID,
				Name:              snap.Name,
				CreatedAt:         time.Unix(int64(snap.CreatedAtUnix), 0),
			})
		}
		return snapshots, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// DownloadSnapshotArchive downloads a project snapshot as a zip archive.
func (s *ProjectsService) DownloadSnapshotArchive(ctx context.Context, projectID, snapshotID string) (io.Reader, error) {
	if projectID == "" {
		return nil, &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}
	if snapshotID == "" {
		return nil, &ValidationError{Field: "snapshot_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.StreamProjectSnapshotArchiveEndpoint(ctx,
		api.StreamProjectSnapshotArchiveEndpointParams{
			ProjectID:         projectID,
			ProjectSnapshotID: snapshotID,
		})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.StreamProjectSnapshotArchiveEndpointOK:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// ListChapterSnapshots returns all snapshots for a chapter.
func (s *ProjectsService) ListChapterSnapshots(ctx context.Context, projectID, chapterID string) ([]*ChapterSnapshot, error) {
	if projectID == "" {
		return nil, &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}
	if chapterID == "" {
		return nil, &ValidationError{Field: "chapter_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetChapterSnapshots(ctx, api.GetChapterSnapshotsParams{
		ProjectID: projectID,
		ChapterID: chapterID,
	})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.ChapterSnapshotsResponseModel:
		snapshots := make([]*ChapterSnapshot, 0, len(r.Snapshots))
		for _, snap := range r.Snapshots {
			snapshots = append(snapshots, &ChapterSnapshot{
				ChapterSnapshotID: snap.ChapterSnapshotID,
				ProjectID:         snap.ProjectID,
				ChapterID:         snap.ChapterID,
				Name:              snap.Name,
				CreatedAt:         time.Unix(int64(snap.CreatedAtUnix), 0),
			})
		}
		return snapshots, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// StreamChapterAudio streams audio from a chapter snapshot.
func (s *ProjectsService) StreamChapterAudio(ctx context.Context, projectID, chapterID, snapshotID string) (io.Reader, error) {
	if projectID == "" {
		return nil, &ValidationError{Field: "project_id", Message: "cannot be empty"}
	}
	if chapterID == "" {
		return nil, &ValidationError{Field: "chapter_id", Message: "cannot be empty"}
	}
	if snapshotID == "" {
		return nil, &ValidationError{Field: "snapshot_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.StreamChapterSnapshotAudio(ctx,
		api.OptBodyStreamChapterAudioV1StudioProjectsProjectIDChaptersChapterIDSnapshotsChapterSnapshotIDStreamPost{},
		api.StreamChapterSnapshotAudioParams{
			ProjectID:         projectID,
			ChapterID:         chapterID,
			ChapterSnapshotID: snapshotID,
		})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.StreamChapterSnapshotAudioOK:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// projectFromAPI converts an API ProjectResponseModel to our Project type.
func projectFromAPI(p *api.ProjectResponseModel) *Project {
	proj := &Project{
		ProjectID:               p.ProjectID,
		Name:                    p.Name,
		DefaultModelID:          p.DefaultModelID,
		DefaultParagraphVoiceID: p.DefaultParagraphVoiceID,
		DefaultTitleVoiceID:     p.DefaultTitleVoiceID,
		CreatedAt:               time.Unix(int64(p.CreateDateUnix), 0),
		CanBeDownloaded:         p.CanBeDownloaded,
		AccessLevel:             string(p.AccessLevel),
	}

	if p.Description.Set && !p.Description.Null {
		proj.Description = p.Description.Value
	}
	if p.Author.Set && !p.Author.Null {
		proj.Author = p.Author.Value
	}
	if p.Language.Set && !p.Language.Null {
		proj.Language = p.Language.Value
	}
	if p.ContentType.Set && !p.ContentType.Null {
		proj.ContentType = p.ContentType.Value
	}
	if p.CoverImageURL.Set && !p.CoverImageURL.Null {
		proj.CoverImageURL = p.CoverImageURL.Value
	}

	return proj
}
