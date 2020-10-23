package client

import (
	"github.com/Kami-no/ward/src/app/client/gitlabclient/clientmock"
	"github.com/Kami-no/ward/src/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/xanzy/go-gitlab"
	"testing"
)

type ClientTest struct {
	suite.Suite

	clientMock *clientmock.ClientMock
}

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
	gitlabClient := NewGitlabClient(&config, c.clientMock)

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
