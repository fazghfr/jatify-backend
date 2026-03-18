//go:build integration

package notion_test

import (
	"context"
	"os"
	"testing"

	notion "job-tracker/internal/notion"
)

func TestIntegration_ListDatabases(t *testing.T) {
	token := os.Getenv("NOTION_TEST_TOKEN")
	if token == "" {
		t.Skip("NOTION_TEST_TOKEN not set")
	}

	c := notion.New(token)
	dbs, err := c.ListDatabases(context.Background())
	if err != nil {
		t.Fatalf("ListDatabases error: %v", err)
	}
	t.Logf("found %d databases", len(dbs))
}

func TestIntegration_QueryDatabase(t *testing.T) {
	token := os.Getenv("NOTION_TEST_TOKEN")
	dbID := os.Getenv("NOTION_TEST_DATABASE_ID")
	if token == "" {
		t.Skip("NOTION_TEST_TOKEN not set")
	}
	if dbID == "" {
		t.Skip("NOTION_TEST_DATABASE_ID not set")
	}

	c := notion.New(token)
	pages, err := c.QueryDatabase(context.Background(), dbID)
	if err != nil {
		t.Fatalf("QueryDatabase error: %v", err)
	}
	t.Logf("found %d pages", len(pages))
}
