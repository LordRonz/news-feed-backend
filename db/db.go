package db

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lordronz/news-feed-backend/model"
)

var DB *sql.DB

func Connect(uri string) error {
	var err error
	DB, err = sql.Open("pgx", uri)
	if err != nil {
		return err
	}
	return DB.Ping()
}

func EncodeCursor(ts int64) string {
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", ts)))
}

func DecodeCursor(encoded string) (int64, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(data), 10, 64)
}

func GetFeeds(ctx context.Context, cursorTs *int64, cursorID *string, limit int) ([]model.Feed, *int64, *string, error) {
	query := `
		SELECT
			f.id, f.content, f.created_at,
			a.id, a.name,
			COALESCE(r.likes, 0), COALESCE(r.haha, 0)
		FROM feeds f
		JOIN authors a ON f.author_id = a.id
		LEFT JOIN reactions r ON f.id = r.feed_id
		WHERE (
			$1::bigint IS NULL OR
			(EXTRACT(EPOCH FROM f.created_at) < $1::bigint) OR
			(EXTRACT(EPOCH FROM f.created_at) = $1::bigint AND f.id < $2::uuid)
		)
		ORDER BY f.created_at DESC, f.id DESC
		LIMIT $3
	`

	var tsVal any = nil
	var idVal any = nil
	if cursorTs != nil {
		tsVal = *cursorTs
	}
	if cursorID != nil {
		idVal = *cursorID
	}

	rows, err := DB.QueryContext(ctx, query, tsVal, idVal, limit+1)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var feeds []model.Feed
	for rows.Next() {
		var f model.Feed
		var createdAt time.Time
		if err := rows.Scan(
			&f.ID, &f.Content, &createdAt,
			&f.Author.ID, &f.Author.Name,
			&f.Reactions.Likes, &f.Reactions.Haha,
		); err != nil {
			return nil, nil, nil, err
		}
		f.CreatedTime = createdAt.Unix()
		feeds = append(feeds, f)
	}

	var nextCursorTs *int64
	var nextCursorID *string
	if len(feeds) > limit {
		t := feeds[limit-1].CreatedTime
		id := feeds[limit-1].ID
		nextCursorTs = &t
		nextCursorID = &id
		feeds = feeds[:limit]
	}

	return feeds, nextCursorTs, nextCursorID, nil
}

func InsertFeed(ctx context.Context, authorID string, content string, createdAt time.Time) (string, error) {
	query := `
		INSERT INTO feeds (author_id, content, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id string
	err := DB.QueryRowContext(ctx, query, authorID, content, createdAt).Scan(&id)
	return id, err
}

func InsertFeedEvent(ctx context.Context, feedID, authorID string, createdAt time.Time) error {
	query := `INSERT INTO feed_events (feed_id, author_id, created_at) VALUES ($1, $2, $3)`
	_, err := DB.ExecContext(ctx, query, feedID, authorID, createdAt)
	return err
}
