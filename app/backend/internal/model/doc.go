package model

type DocTocItem struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type DocTocSection struct {
	Title string       `json:"title"`
	Items []DocTocItem `json:"items"`
}
