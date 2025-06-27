package model

type Feed struct {
	ID          string    `json:"id"`
	Author      Author    `json:"author"`
	Content     string    `json:"content"`
	Reactions   Reactions `json:"reactions"`
	CreatedTime int64     `json:"created_time"` // UNIX timestamp
}

type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Reactions struct {
	Likes int `json:"likes"`
	Haha  int `json:"haha"`
}

type Pagination struct {
	Size       int    `json:"size"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type FeedResponse struct {
	Pagination Pagination `json:"pagination"`
	Results    []Feed     `json:"results"`
}