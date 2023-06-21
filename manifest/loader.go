package manifest

type FileRecord struct {
	Path      string `json:"path"`
	QuickSHA1 string `json:"quickSHA1"`
	Size      uint64 `json:"size"`
}

type ScriptRecord struct {
	Path  string       `json:"path"`
	Files []FileRecord `json:"files"`
}

type Manifest struct {
	Sources []ScriptRecord `json:"sources"`
}
