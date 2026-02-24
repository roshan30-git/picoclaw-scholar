// Package gdrive provides Google Drive integration for StudyClaw.
// Used for: storing textbooks, PDFs, lecture slides, and shared study materials.
// The agent can list, search, and download files from a designated Drive folder.
package gdrive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	// StudyClawFolderName is the root Drive folder StudyClaw will use.
	StudyClawFolderName = "StudyClaw Books"
	// CredentialsFile is the path to the Google OAuth2 credentials JSON.
	CredentialsFile = "~/.studyclaw/google_credentials.json"
	// TokenFile stores the OAuth2 token after first login.
	TokenFile = "~/.studyclaw/google_token.json"
)

// Client wraps the Google Drive service.
type Client struct {
	svc      *drive.Service
	rootID   string // ID of the StudyClaw Books folder
}

// New authenticates and returns a new Google Drive client.
// On first run, it opens a browser for OAuth2 consent.
func New(ctx context.Context) (*Client, error) {
	credBytes, err := os.ReadFile(expandHome(CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w (see README for setup)", err)
	}

	config, err := google.ConfigFromJSON(credBytes, drive.DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	token, err := loadToken(expandHome(TokenFile))
	if err != nil {
		// First run: generate token via browser
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(expandHome(TokenFile), token)
	}

	ts := config.TokenSource(ctx, token)
	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("drive service: %w", err)
	}

	c := &Client{svc: svc}

	// Find or create the StudyClaw Books folder
	c.rootID, err = c.ensureFolder(ctx, StudyClawFolderName)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ListBooks returns all PDF/DOCX files in the StudyClaw Books folder.
func (c *Client) ListBooks(ctx context.Context) ([]*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and trashed = false and (mimeType='application/pdf' or mimeType='application/msword')", c.rootID)
	res, err := c.svc.Files.List().
		Q(query).
		Fields("files(id, name, size, modifiedTime)").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	return res.Files, nil
}

// DownloadBook downloads a file by Drive ID to the local workspace.
func (c *Client) DownloadBook(ctx context.Context, fileID, destDir string) (string, error) {
	meta, err := c.svc.Files.Get(fileID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("get file meta: %w", err)
	}

	res, err := c.svc.Files.Get(fileID).Context(ctx).Download()
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer res.Body.Close()

	destPath := filepath.Join(destDir, meta.Name)
	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, res.Body); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return destPath, nil
}

// SearchBooks searches for a book by name keyword in the StudyClaw folder.
func (c *Client) SearchBooks(ctx context.Context, keyword string) ([]*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and name contains '%s' and trashed = false", c.rootID, keyword)
	res, err := c.svc.Files.List().Q(query).Fields("files(id, name, size)").Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return res.Files, nil
}

// ensureFolder finds or creates a folder by name at Drive root.
func (c *Client) ensureFolder(ctx context.Context, name string) (string, error) {
	q := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and trashed=false", name)
	res, err := c.svc.Files.List().Q(q).Fields("files(id)").Context(ctx).Do()
	if err != nil {
		return "", err
	}
	if len(res.Files) > 0 {
		return res.Files[0].Id, nil
	}
	// Create it
	folder, err := c.svc.Files.Create(&drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create folder: %w", err)
	}
	return folder.Id, nil
}

func loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	if err := newJSONDecoder(f).Decode(token); err != nil {
		return nil, err
	}
	return token, nil
}

func saveToken(path string, token *oauth2.Token) {
	f, _ := os.Create(path)
	defer f.Close()
	newJSONEncoder(f).Encode(token)
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("📂 Open this URL in your browser to link Google Drive:\n%s\n\nPaste the code here: ", authURL)
	var code string
	fmt.Scan(&code)
	return config.Exchange(context.Background(), code)
}

func expandHome(path string) string {
	home, _ := os.UserHomeDir()
	if len(path) > 1 && path[:2] == "~/" {
		return filepath.Join(home, path[2:])
	}
	return path
}

// newJSONDecoder / newJSONEncoder are thin wrappers to avoid importing encoding/json at top level.
func newJSONDecoder(r io.Reader) interface{ Decode(v any) error } {
	import_json_decoder, _ := r.(interface{ Decode(v any) error })
	return import_json_decoder
}
func newJSONEncoder(w io.Writer) interface{ Encode(v any) error } {
	import_json_encoder, _ := w.(interface{ Encode(v any) error })
	return import_json_encoder
}
