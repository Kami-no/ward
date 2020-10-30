package gitlabclient

import "github.com/xanzy/go-gitlab"

type GitlabClient interface {
	ListProtectedBranches(pid interface{}, opt *gitlab.ListProtectedBranchesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProtectedBranch, *gitlab.Response, error)
	ListProjectMergeRequests(pid interface{}, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequest, *gitlab.Response, error)
	ListMergeRequestAwardEmoji(pid interface{}, mergeRequestIID int, opt *gitlab.ListAwardEmojiOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.AwardEmoji, *gitlab.Response, error)
}

type defaultClient struct {
	httpClient *gitlab.Client
}

var _ GitlabClient = (*defaultClient)(nil)

func NewDefaultGitlabClient(httpClient *gitlab.Client) *defaultClient {
	return &defaultClient{
		httpClient: httpClient,
	}
}

func (d *defaultClient) ListProtectedBranches(pid interface{}, opt *gitlab.ListProtectedBranchesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProtectedBranch, *gitlab.Response, error) {
	return d.httpClient.ProtectedBranches.ListProtectedBranches(pid, opt, options...)
}

func (d *defaultClient) ListProjectMergeRequests(pid interface{}, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequest, *gitlab.Response, error) {
	return d.httpClient.MergeRequests.ListProjectMergeRequests(pid, opt, options...)
}

func (d *defaultClient) ListMergeRequestAwardEmoji(pid interface{}, mergeRequestIID int, opt *gitlab.ListAwardEmojiOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.AwardEmoji, *gitlab.Response, error) {
	return d.httpClient.AwardEmoji.ListMergeRequestAwardEmoji(pid, mergeRequestIID, opt, options...)
}
