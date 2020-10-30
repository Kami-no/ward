package clientmock

import (
	"github.com/Kami-no/ward/src/app/client/gitlabclient"
	"github.com/stretchr/testify/mock"
	"github.com/xanzy/go-gitlab"
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
