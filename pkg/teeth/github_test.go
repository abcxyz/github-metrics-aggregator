package teeth

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v56/github"
	"github.com/sethvargo/go-retry"
)

const (
	testPRNum     = 1
	testPRComment = "foo"
)

func TestCreateInvocationCommentWithRetry(t *testing.T) {
	t.Parallel()

	retryFunc = retry.NewConstant
	retryMaxAttempts = 1

	tests := []struct {
		name          string
		commentPRFunc func(ctx context.Context, owner, repo string, prNumber int, comment string) (*github.Response, error)
		wantErr       bool
	}{
		{
			name: "status_ok",
			commentPRFunc: func(_ context.Context, _, _ string, _ int, _ string) (*github.Response, error) {
				return &github.Response{Response: &http.Response{
					StatusCode: http.StatusCreated,
					Body:       http.NoBody,
				}}, nil
			},
		},
		{
			name: "status_not_retryable",
			commentPRFunc: func(_ context.Context, _, _ string, _ int, _ string) (*github.Response, error) {
				return &github.Response{Response: &http.Response{
					StatusCode: http.StatusGone,
					Body:       http.NoBody,
				}}, nil
			},
			wantErr: true,
		},
		{
			name: "status_always_retryable",
			commentPRFunc: func(_ context.Context, _, _ string, _ int, _ string) (*github.Response, error) {
				return &github.Response{Response: &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       http.NoBody,
				}}, nil
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := &GitHub{
				config:        &GHConfig{},
				commentPRFunc: tc.commentPRFunc,
			}
			err := client.CreateInvocationCommentWithRetry(context.Background(), testPRNum, testPRComment)
			if tc.wantErr != (err != nil) {
				t.Errorf("CreateInvocationCommentWithRetry returned error: %v; want error? %t", err, tc.wantErr)
			}
		})
	}
}
