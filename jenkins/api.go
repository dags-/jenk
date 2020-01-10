package jenkins

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

type Git struct {
	Revision  *GitRevision `json:"lastBuiltRevision"`
	RemoteURL []string     `json:"remoteUrls"`
}

type GitRevision struct {
	SHA1 string `json:"SHA1"`
}
