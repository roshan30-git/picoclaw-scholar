package classroom

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

const (
	CredentialsFile = "~/.studyclaw/google_credentials.json"
	TokenFile       = "~/.studyclaw/google_classroom_token.json"
)

type Client struct {
	svc *classroom.Service
}

// New creates a new Google Classroom client, sharing credentials with Drive.
func New(ctx context.Context) (*Client, error) {
	credBytes, err := os.ReadFile(expandHome(CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	config, err := google.ConfigFromJSON(credBytes, classroom.ClassroomCoursesReadonlyScope, classroom.ClassroomCourseworkMeReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	token, err := loadToken(expandHome(TokenFile))
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(expandHome(TokenFile), token)
	}

	ts := config.TokenSource(ctx, token)
	svc, err := classroom.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("classroom service: %w", err)
	}

	return &Client{svc: svc}, nil
}

// ListCourses gets all active courses the student is enrolled in.
func (c *Client) ListCourses(ctx context.Context) ([]*classroom.Course, error) {
	res, err := c.svc.Courses.List().StudentId("me").CourseStates("ACTIVE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	return res.Courses, nil
}

// ListAssignments fetches assignments for a particular course.
func (c *Client) ListAssignments(ctx context.Context, courseID string) ([]*classroom.CourseWork, error) {
	res, err := c.svc.Courses.CourseWork.List(courseID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list coursework: %w", err)
	}
	return res.CourseWork, nil
}

// Below are standard OAuth helpers mirroring the gdrive module

func loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, err
	}
	return token, nil
}

func saveToken(path string, token *oauth2.Token) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Warning: Unable to cache oauth token: %v\n", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("🎓 Open this URL to link Google Classroom:\n%s\n\nPaste code: ", authURL)
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
