package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/lordronz/news-feed-backend/db"
	"github.com/lordronz/news-feed-backend/model"
	"github.com/lordronz/news-feed-backend/queue"
	"github.com/lordronz/news-feed-backend/util"
)

func GetFeeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	limit := 10
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	var cursorTs *int64
	var cursorID *string
	if c := q.Get("cursor"); c != "" {
		if ts, id, err := util.DecodeCursor(c); err == nil {
			cursorTs = &ts
			cursorID = &id
		}
	}

	feeds, nextCursorTs, nextCursorId, err := db.GetFeeds(ctx, cursorTs, cursorID, limit)
	if err != nil {
		http.Error(w, "Failed to load feeds", http.StatusInternalServerError)
		return
	}

	response := model.FeedResponse{
		Pagination: model.Pagination{
			Size:       limit,
			NextCursor: "",
		},
		Results: feeds,
	}
	if nextCursorTs != nil && nextCursorId != nil {
		response.Pagination.NextCursor = util.EncodeCursor(*nextCursorTs, *nextCursorId)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content  string `json:"content"`
		AuthorID string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Content == "" || req.AuthorID == "" {
		http.Error(w, "Missing content or author_id", http.StatusBadRequest)
		return
	}

	createdAt := time.Now()
	feedID, err := db.InsertFeed(r.Context(), req.AuthorID, req.Content, createdAt)
	if err != nil {
		log.Printf("❌ Failed to insert feed: %v", err)
		http.Error(w, "Failed to create feed", http.StatusInternalServerError)
		return
	}

	_ = queue.PublishFeedCreated(queue.FeedEvent{
		ID:      feedID,
		Content: req.Content,
		Author:  req.AuthorID,
		Time:    createdAt.Unix(),
	})

	resp := model.Feed{
		ID:      feedID,
		Content: req.Content,
		Author:  model.Author{ID: req.AuthorID, Name: ""}, // Name not loaded here
		Reactions: model.Reactions{
			Likes: 0,
			Haha:  0,
		},
		CreatedTime: createdAt.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
