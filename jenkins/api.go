package jenkins

var emptyBuild = &Build{
	Number:    -1,
	Result:    "FAIL",
	Artifacts: []*Artifact{},
}

type Job struct {
	Builds []*BuildMeta `json:"builds"`
}

type BuildMeta struct {
	URL string `json:"url"`
}

type Build struct {
	URL       string      `json:"url"`
	Number    int         `json:"number"`
	Result    string      `json:"result"`
	Timestamp int64       `json:"timestamp"`
	Artifacts []*Artifact `json:"artifacts"`
}

type InternalBuild struct {
	Number    int         `json:"number"`
	Result    string      `json:"result"`
	Timestamp int64       `json:"timestamp"`
	Artifacts []*Artifact `json:"artifacts"`
}

type Artifact struct {
	FileName string `json:"fileName"`
	Path     string `json:"relativePath"`
}
