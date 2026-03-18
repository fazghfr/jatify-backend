package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

type Client struct {
	inner *notionapi.Client
}

func New(accessToken string) *Client {
	return &Client{inner: notionapi.NewClient(notionapi.Token(accessToken))}
}

// QueryDatabase paginates through all pages in the database and returns them all.
func (c *Client) QueryDatabase(ctx context.Context, databaseID string) ([]notionapi.Page, error) {
	var pages []notionapi.Page
	var cursor notionapi.Cursor
	for {
		resp, err := c.inner.Database.Query(ctx, notionapi.DatabaseID(databaseID), &notionapi.DatabaseQueryRequest{
			StartCursor: cursor,
			PageSize:    100,
		})
		if err != nil {
			return nil, err
		}
		pages = append(pages, resp.Results...)
		if !resp.HasMore {
			break
		}
		cursor = notionapi.Cursor(resp.NextCursor)
	}
	return pages, nil
}

// ListDatabases returns all databases the integration can access.
func (c *Client) ListDatabases(ctx context.Context) ([]notionapi.Database, error) {
	var databases []notionapi.Database
	var cursor notionapi.Cursor
	for {
		resp, err := c.inner.Search.Do(ctx, &notionapi.SearchRequest{
			Filter:      notionapi.SearchFilter{Value: "database", Property: "object"},
			StartCursor: cursor,
			PageSize:    100,
		})
		if err != nil {
			return nil, err
		}
		for _, result := range resp.Results {
			if db, ok := result.(*notionapi.Database); ok {
				databases = append(databases, *db)
			}
		}
		if !resp.HasMore {
			break
		}
		cursor = notionapi.Cursor(resp.NextCursor)
	}
	return databases, nil
}
