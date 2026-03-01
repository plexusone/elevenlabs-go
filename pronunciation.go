package elevenlabs

import (
	"context"
	"io"
	"time"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// PronunciationService handles pronunciation dictionary operations.
// Pronunciation dictionaries help ensure correct pronunciation of
// technical terms, names, and domain-specific vocabulary.
type PronunciationService struct {
	client *Client
}

// PronunciationDictionary represents a pronunciation dictionary.
type PronunciationDictionary struct {
	// ID is the unique identifier.
	ID string

	// Name is the display name.
	Name string

	// Description is the dictionary description.
	Description string

	// LatestVersionID is the ID of the latest version.
	LatestVersionID string

	// RulesCount is the number of rules in the latest version.
	RulesCount int

	// CreatedBy is the user ID who created the dictionary.
	CreatedBy string

	// CreatedAt is when the dictionary was created.
	CreatedAt time.Time
}

// PronunciationDictionaryListResponse contains the list result.
type PronunciationDictionaryListResponse struct {
	// Dictionaries is the list of pronunciation dictionaries.
	Dictionaries []*PronunciationDictionary

	// HasMore indicates if there are more items to fetch.
	HasMore bool

	// NextCursor is the cursor for pagination.
	NextCursor string
}

// PronunciationDictionaryListOptions contains options for listing.
type PronunciationDictionaryListOptions struct {
	// PageSize is the number of items per page (max 100).
	PageSize int

	// Cursor is the pagination cursor.
	Cursor string
}

// List returns all pronunciation dictionaries.
func (s *PronunciationService) List(ctx context.Context, opts *PronunciationDictionaryListOptions) (*PronunciationDictionaryListResponse, error) {
	params := api.GetPronunciationDictionariesMetadataParams{}

	if opts != nil {
		if opts.PageSize > 0 {
			params.PageSize = api.NewOptInt(opts.PageSize)
		}
		if opts.Cursor != "" {
			params.Cursor = api.NewOptNilString(opts.Cursor)
		}
	}

	resp, err := s.client.apiClient.GetPronunciationDictionariesMetadata(ctx, params)
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.GetPronunciationDictionariesMetadataResponseModel:
		result := &PronunciationDictionaryListResponse{
			HasMore:      r.HasMore,
			Dictionaries: make([]*PronunciationDictionary, 0, len(r.PronunciationDictionaries)),
		}

		if r.NextCursor.Set && !r.NextCursor.Null {
			result.NextCursor = r.NextCursor.Value
		}

		for _, d := range r.PronunciationDictionaries {
			dict := &PronunciationDictionary{
				ID:              d.ID,
				Name:            d.Name,
				LatestVersionID: d.LatestVersionID,
				RulesCount:      d.LatestVersionRulesNum,
				CreatedBy:       d.CreatedBy,
				CreatedAt:       time.Unix(int64(d.CreationTimeUnix), 0),
			}
			if d.Description.Set && !d.Description.Null {
				dict.Description = d.Description.Value
			}
			result.Dictionaries = append(result.Dictionaries, dict)
		}

		return result, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Get returns a pronunciation dictionary by ID.
func (s *PronunciationService) Get(ctx context.Context, dictionaryID string) (*PronunciationDictionary, error) {
	if dictionaryID == "" {
		return nil, &ValidationError{Field: "dictionary_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetPronunciationDictionaryMetadata(ctx, api.GetPronunciationDictionaryMetadataParams{
		PronunciationDictionaryID: dictionaryID,
	})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.GetPronunciationDictionaryMetadataResponseModel:
		dict := &PronunciationDictionary{
			ID:              r.ID,
			Name:            r.Name,
			LatestVersionID: r.LatestVersionID,
			RulesCount:      r.LatestVersionRulesNum,
			CreatedBy:       r.CreatedBy,
			CreatedAt:       time.Unix(int64(r.CreationTimeUnix), 0),
		}
		if r.Description.Set && !r.Description.Null {
			dict.Description = r.Description.Value
		}
		return dict, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// CreatePronunciationDictionaryRequest contains options for creating a pronunciation dictionary.
type CreatePronunciationDictionaryRequest struct {
	// Name is the name of the dictionary (required).
	Name string

	// Description is an optional description.
	Description string

	// PLSContent is the PLS (Pronunciation Lexicon Specification) XML content.
	// Use this to provide pronunciation rules directly.
	// You can generate this from PronunciationRules using ToPLSString().
	PLSContent string

	// Rules is a convenient alternative to PLSContent.
	// If provided, it will be converted to PLS format automatically.
	// If both PLSContent and Rules are provided, PLSContent takes precedence.
	Rules PronunciationRules

	// Language is the language code for the rules (default: "en-US").
	// Only used when Rules is provided.
	Language string
}

// Create creates a new pronunciation dictionary.
//
// Example with rules:
//
//	dict, err := client.Pronunciation().Create(ctx, &CreatePronunciationDictionaryRequest{
//	    Name: "Tech Terms",
//	    Rules: elevenlabs.RulesFromMap(map[string]string{
//	        "ADK":     "Agent Development Kit",
//	        "kubectl": "kube control",
//	    }),
//	})
//
// Example with PLS content:
//
//	dict, err := client.Pronunciation().Create(ctx, &CreatePronunciationDictionaryRequest{
//	    Name:       "Tech Terms",
//	    PLSContent: plsXMLString,
//	})
func (s *PronunciationService) Create(ctx context.Context, req *CreatePronunciationDictionaryRequest) (*PronunciationDictionary, error) {
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "cannot be empty"}
	}

	body := &api.BodyAddAPronunciationDictionaryV1PronunciationDictionariesAddFromFilePostMultipart{
		Name: req.Name,
	}

	if req.Description != "" {
		body.Description = api.NewOptNilString(req.Description)
	}

	// Handle PLS content
	plsContent := req.PLSContent
	if plsContent == "" && len(req.Rules) > 0 {
		// Generate PLS from rules
		lang := req.Language
		if lang == "" {
			lang = "en-US"
		}
		generated, err := req.Rules.ToPLSString(lang)
		if err != nil {
			return nil, err
		}
		plsContent = generated
	}

	if plsContent != "" {
		body.File = api.NewOptNilString(plsContent)
	}

	resp, err := s.client.apiClient.AddFromFile(ctx, body, api.AddFromFileParams{})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.AddPronunciationDictionaryResponseModel:
		dict := &PronunciationDictionary{
			ID:              r.ID,
			Name:            r.Name,
			LatestVersionID: r.VersionID,
			RulesCount:      r.VersionRulesNum,
			CreatedBy:       r.CreatedBy,
			CreatedAt:       time.Unix(int64(r.CreationTimeUnix), 0),
		}
		if r.Description.Set && !r.Description.Null {
			dict.Description = r.Description.Value
		}
		return dict, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// CreateFromJSON creates a pronunciation dictionary from a JSON rules file.
//
// Example JSON file:
//
//	[
//	  {"grapheme": "ADK", "alias": "Agent Development Kit"},
//	  {"grapheme": "kubectl", "alias": "kube control"}
//	]
func (s *PronunciationService) CreateFromJSON(ctx context.Context, name, jsonFilePath string) (*PronunciationDictionary, error) {
	rules, err := LoadRulesFromJSON(jsonFilePath)
	if err != nil {
		return nil, err
	}

	return s.Create(ctx, &CreatePronunciationDictionaryRequest{
		Name:  name,
		Rules: rules,
	})
}

// CreateFromMap creates a pronunciation dictionary from a simple map.
// All entries are treated as alias substitutions (text replacements).
//
// Example:
//
//	dict, err := client.Pronunciation().CreateFromMap(ctx, "Tech Terms", map[string]string{
//	    "ADK":     "Agent Development Kit",
//	    "kubectl": "kube control",
//	    "API":     "A P I",
//	})
func (s *PronunciationService) CreateFromMap(ctx context.Context, name string, rules map[string]string) (*PronunciationDictionary, error) {
	return s.Create(ctx, &CreatePronunciationDictionaryRequest{
		Name:  name,
		Rules: RulesFromMap(rules),
	})
}

// RemoveRules removes pronunciation rules from a dictionary.
// The ruleStrings should be the original text strings to remove.
func (s *PronunciationService) RemoveRules(ctx context.Context, dictionaryID string, ruleStrings []string) error {
	if dictionaryID == "" {
		return &ValidationError{Field: "dictionary_id", Message: "cannot be empty"}
	}
	if len(ruleStrings) == 0 {
		return &ValidationError{Field: "rule_strings", Message: "cannot be empty"}
	}

	body := &api.BodyRemoveRulesFromThePronunciationDictionaryV1PronunciationDictionariesPronunciationDictionaryIDRemoveRulesPost{
		RuleStrings: ruleStrings,
	}

	_, err := s.client.apiClient.RemoveRules(ctx, body, api.RemoveRulesParams{
		PronunciationDictionaryID: dictionaryID,
	})
	return err
}

// Rename renames a pronunciation dictionary.
func (s *PronunciationService) Rename(ctx context.Context, dictionaryID, newName string) error {
	if dictionaryID == "" {
		return &ValidationError{Field: "dictionary_id", Message: "cannot be empty"}
	}
	if newName == "" {
		return &ValidationError{Field: "new_name", Message: "cannot be empty"}
	}

	body := api.BodyUpdatePronunciationDictionaryV1PronunciationDictionariesPronunciationDictionaryIDPatch{
		Name: api.NewOptString(newName),
	}

	_, err := s.client.apiClient.PatchPronunciationDictionary(ctx,
		api.NewOptBodyUpdatePronunciationDictionaryV1PronunciationDictionariesPronunciationDictionaryIDPatch(body),
		api.PatchPronunciationDictionaryParams{
			PronunciationDictionaryID: dictionaryID,
		})
	return err
}

// Archive archives a pronunciation dictionary.
func (s *PronunciationService) Archive(ctx context.Context, dictionaryID string) error {
	if dictionaryID == "" {
		return &ValidationError{Field: "dictionary_id", Message: "cannot be empty"}
	}

	body := api.BodyUpdatePronunciationDictionaryV1PronunciationDictionariesPronunciationDictionaryIDPatch{
		Archived: api.NewOptBool(true),
	}

	_, err := s.client.apiClient.PatchPronunciationDictionary(ctx,
		api.NewOptBodyUpdatePronunciationDictionaryV1PronunciationDictionariesPronunciationDictionaryIDPatch(body),
		api.PatchPronunciationDictionaryParams{
			PronunciationDictionaryID: dictionaryID,
		})
	return err
}

// GetVersionPLS returns the PLS (Pronunciation Lexicon Specification) XML file
// for a specific version of a pronunciation dictionary.
//
// The returned io.Reader contains the XML content that can be saved to a file
// or parsed directly.
//
// Example:
//
//	pls, err := client.Pronunciation().GetVersionPLS(ctx, dictionaryID, versionID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Save to file
//	f, _ := os.Create("dictionary.pls")
//	io.Copy(f, pls)
func (s *PronunciationService) GetVersionPLS(ctx context.Context, dictionaryID, versionID string) (io.Reader, error) {
	if dictionaryID == "" {
		return nil, &ValidationError{Field: "dictionary_id", Message: "cannot be empty"}
	}
	if versionID == "" {
		return nil, &ValidationError{Field: "version_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetPronunciationDictionaryVersionPls(ctx, api.GetPronunciationDictionaryVersionPlsParams{
		DictionaryID: dictionaryID,
		VersionID:    versionID,
	})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.GetPronunciationDictionaryVersionPlsOKHeaders:
		return r.Response.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// DownloadLatestPLS downloads the PLS file for the latest version of a dictionary.
// This is a convenience method that first gets the dictionary metadata to find
// the latest version ID, then downloads that version.
func (s *PronunciationService) DownloadLatestPLS(ctx context.Context, dictionaryID string) (io.Reader, error) {
	// Get dictionary to find latest version
	dict, err := s.Get(ctx, dictionaryID)
	if err != nil {
		return nil, err
	}

	return s.GetVersionPLS(ctx, dictionaryID, dict.LatestVersionID)
}
