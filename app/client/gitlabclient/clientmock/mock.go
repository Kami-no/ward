package clientmock

import (
	"github.com/stretchr/testify/mock"
	"github.com/xanzy/go-gitlab"

	"github.com/Kami-no/ward/app/client/gitlabclient"
)

type ClientMock struct {
	mock.Mock
}

var _ gitlabclient.GitlabClient = (*ClientMock)(nil)

func (c *ClientMock) Reset() {
	c.Mock = mock.Mock{}
}

func (c *ClientMock) ListProtectedBranches(pid interface{}, opt *gitlab.ListProtectedBranchesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.ProtectedBranch, *gitlab.Response, error) {
	arguments := c.Called(pid, opt, options)
	return arguments.Get(0).([]*gitlab.ProtectedBranch), arguments.Get(1).(*gitlab.Response), arguments.Error(2)
}

func (c *ClientMock) ListProjectMergeRequests(pid interface{}, opt *gitlab.ListProjectMergeRequestsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequest, *gitlab.Response, error) {
	arguments := c.Called(pid, opt, options)
	return arguments.Get(0).([]*gitlab.MergeRequest), arguments.Get(1).(*gitlab.Response), arguments.Error(2)
}

func (c *ClientMock) ListMergeRequestAwardEmoji(pid interface{}, mergeRequestIID int, opt *gitlab.ListAwardEmojiOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.AwardEmoji, *gitlab.Response, error) {
	arguments := c.Called(pid, mergeRequestIID, opt, options)
	return arguments.Get(0).([]*gitlab.AwardEmoji), arguments.Get(1).(*gitlab.Response), arguments.Error(2)
}

func (c *ClientMock) GetProject(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	arguments := c.Called(pid, opt, options)
	return arguments.Get(0).(*gitlab.Project), arguments.Get(1).(*gitlab.Response), arguments.Error(2)
}

func (c *ClientMock) ListBranches(pid interface{}, opt *gitlab.ListBranchesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Branch, *gitlab.Response, error) {
	arguments := c.Called(pid, opt, options)
	return arguments.Get(0).([]*gitlab.Branch), arguments.Get(1).(*gitlab.Response), arguments.Error(2)
}
