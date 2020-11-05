package client

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/xanzy/go-gitlab"

	"github.com/Kami-no/ward/app/client/gitlabclient/clientmock"
	"github.com/Kami-no/ward/app/ldap"
	"github.com/Kami-no/ward/config"
)

type ClientTest struct {
	suite.Suite

	clientMock *clientmock.ClientMock
}

type ldapStub struct {
}

func (l ldapStub) Check(_ string) bool {
	return false
}

func (l ldapStub) ListMails(_ []string) []string {
	return nil
}

var _ ldap.Service = (*ldapStub)(nil)

func TestGitlabClient(t *testing.T) {
	suite.Run(t, &ClientTest{})
}

func (c *ClientTest) SetupSuite() {
	c.clientMock = &clientmock.ClientMock{}
}

func (c *ClientTest) TestSuccessfulProcessing() {
	projectID := 1
	config := config.Config{
		Projects: map[int]*config.Project{
			projectID: {
				Teams: map[string][]string{
					"team1": {"a", "b", "c"},
				},
				Votes: 2,
			},
		},
	}

	defer c.clientMock.Reset()
	gitlabClient := NewGitlabClient(&config, c.clientMock, ldapStub{})

	var response *gitlab.Response

	branches := []*gitlab.ProtectedBranch{
		{
			Name: "protectedBranch",
		},
	}
	c.clientMock.
		On("ListProtectedBranches", projectID, mock.Anything, mock.Anything).
		Return(branches, response, nil)

	mrIID := 234
	mrs := []*gitlab.MergeRequest{
		{
			IID:          mrIID,
			TargetBranch: "protectedBranch",
			WebURL:       "webUrl",
		},
	}
	c.clientMock.
		On("ListProjectMergeRequests", projectID, mock.Anything, mock.Anything).
		Return(mrs, response, nil)

	var awards []*gitlab.AwardEmoji

	c.clientMock.
		On("ListMergeRequestAwardEmoji", projectID, mrIID, mock.Anything, mock.Anything).
		Return(awards, response, nil)

	requests, err := gitlabClient.CheckPrjRequests(config.Projects, "merged")
	c.NoError(err)
	c.NotEmpty(requests[projectID])
	c.NotEmpty(requests[projectID].MR)
	c.Equal("webUrl", requests[projectID].MR[mrIID].Path)
	c.Equal(false, requests[projectID].MR[mrIID].Awards.Like)
}

func (c *ClientTest) TestEmptyMergeRequestsProcessing() {
	projectID := 1
	config := config.Config{
		Projects: map[int]*config.Project{
			projectID: {
				Teams: map[string][]string{
					"team1": {"a", "b", "c"},
				},
				Votes: 2,
			},
		},
	}
	defer c.clientMock.Reset()
	gitlabClient := NewGitlabClient(&config, c.clientMock, ldapStub{})

	var response *gitlab.Response

	branches := []*gitlab.ProtectedBranch{
		{
			Name: "protectedBranch",
		},
	}
	c.clientMock.
		On("ListProtectedBranches", projectID, mock.Anything, mock.Anything).
		Return(branches, response, nil)

	mrIID := 234
	var mrs []*gitlab.MergeRequest
	c.clientMock.
		On("ListProjectMergeRequests", projectID, mock.Anything, mock.Anything).
		Return(mrs, response, nil)

	var awards []*gitlab.AwardEmoji

	c.clientMock.
		On("ListMergeRequestAwardEmoji", projectID, mrIID, mock.Anything, mock.Anything).
		Return(awards, response, nil)

	requests, err := gitlabClient.CheckPrjRequests(config.Projects, "merged")
	c.NoError(err)
	c.Empty(requests[projectID])
}
