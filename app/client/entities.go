package client

type MrProject struct {
	MR map[int]MergeRequest
}

type MergeRequest struct {
	Name     string
	Path     string
	MergedBy string
	Awards   struct {
		Like         bool
		Dislike      bool
		Ready        int
		NotReady     int
		NonCompliant int
	}
}

type DeadBranch struct {
	Author string
	Age    int
}

type DeadProject struct {
	Name     string
	URL      string
	Owners   []string
	Branches map[string]DeadBranch
}

type DeadAuthor struct {
	Name     string
	Branches map[int][]string
	Projects map[int]DeadProject
}

type DeadResults struct {
	Projects map[int]DeadProject
	Authors  map[string]DeadAuthor
}
