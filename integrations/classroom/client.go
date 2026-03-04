package classroom

import (
	"context"
	"fmt"

	"github.com/roshan30-git/picoclaw-scholar/pkg/auth"
	"golang.org/x/oauth2"
	"google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

const (
// TokenFile is legacy and should be removed if possible, but keeping for compatibility if needed.
)

type Client struct {
	svc *classroom.Service
}

// New creates a new Google Classroom client using the centralized auth store.
func New(ctx context.Context) (*Client, error) {
	cred, err := auth.GetCredential("google-antigravity")
	if err != nil || cred == nil {
		return nil, fmt.Errorf("google classroom not connected. please run setup and connect google accounts")
	}

	authCfg := auth.GoogleAntigravityOAuthConfig()
	config := &oauth2.Config{
		ClientID:     authCfg.ClientID,
		ClientSecret: authCfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authCfg.Issuer + "/auth",
			TokenURL: authCfg.TokenURL,
		},
		Scopes: []string{
			classroom.ClassroomCoursesReadonlyScope,
			classroom.ClassroomCourseworkMeReadonlyScope,
		},
	}

	token := &oauth2.Token{
		AccessToken:  cred.AccessToken,
		RefreshToken: cred.RefreshToken,
		Expiry:       cred.ExpiresAt,
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
