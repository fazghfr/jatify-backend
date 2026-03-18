package notion

import (
	"testing"

	"github.com/jomei/notionapi"
)

// --- ExtractTitle ---

func TestExtractTitle_WithTitle(t *testing.T) {
	prop := &notionapi.TitleProperty{
		Title: []notionapi.RichText{{PlainText: "Acme Corp"}},
	}
	if got := ExtractTitle(prop); got != "Acme Corp" {
		t.Errorf("got %q, want %q", got, "Acme Corp")
	}
}

func TestExtractTitle_WithWrongType(t *testing.T) {
	prop := &notionapi.RichTextProperty{
		RichText: []notionapi.RichText{{PlainText: "text"}},
	}
	if got := ExtractTitle(prop); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

func TestExtractTitle_Empty(t *testing.T) {
	prop := &notionapi.TitleProperty{Title: []notionapi.RichText{}}
	if got := ExtractTitle(prop); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// --- ExtractRichText ---

func TestExtractRichText_WithContent(t *testing.T) {
	prop := &notionapi.RichTextProperty{
		RichText: []notionapi.RichText{{PlainText: "Software Engineer"}},
	}
	if got := ExtractRichText(prop); got != "Software Engineer" {
		t.Errorf("got %q, want %q", got, "Software Engineer")
	}
}

func TestExtractRichText_Empty(t *testing.T) {
	prop := &notionapi.RichTextProperty{RichText: []notionapi.RichText{}}
	if got := ExtractRichText(prop); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// --- ExtractSelect ---

func TestExtractSelect_WithName(t *testing.T) {
	prop := &notionapi.SelectProperty{
		Select: notionapi.Option{Name: "Applied"},
	}
	if got := ExtractSelect(prop); got != "Applied" {
		t.Errorf("got %q, want %q", got, "Applied")
	}
}

func TestExtractSelect_EmptyName(t *testing.T) {
	prop := &notionapi.SelectProperty{Select: notionapi.Option{}}
	if got := ExtractSelect(prop); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// --- ExtractMultiSelectLast ---

func TestExtractMultiSelectLast_Single(t *testing.T) {
	prop := &notionapi.MultiSelectProperty{
		MultiSelect: []notionapi.Option{{Name: "Applied"}},
	}
	if got := ExtractMultiSelectLast(prop); got != "Applied" {
		t.Errorf("got %q, want %q", got, "Applied")
	}
}

func TestExtractMultiSelectLast_Multiple(t *testing.T) {
	prop := &notionapi.MultiSelectProperty{
		MultiSelect: []notionapi.Option{{Name: "Applied"}, {Name: "Ghosted"}},
	}
	if got := ExtractMultiSelectLast(prop); got != "Ghosted" {
		t.Errorf("got %q, want %q", got, "Ghosted")
	}
}

func TestExtractMultiSelectLast_Empty(t *testing.T) {
	prop := &notionapi.MultiSelectProperty{MultiSelect: []notionapi.Option{}}
	if got := ExtractMultiSelectLast(prop); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

func TestExtractMultiSelectLast_WrongType(t *testing.T) {
	prop := &notionapi.SelectProperty{Select: notionapi.Option{Name: "Applied"}}
	if got := ExtractMultiSelectLast(prop); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// --- MapPage ---

func TestMapPage_AllFields(t *testing.T) {
	page := notionapi.Page{
		ID: "page-123",
		Properties: notionapi.Properties{
			"Company":      &notionapi.TitleProperty{Title: []notionapi.RichText{{PlainText: "Acme"}}},
			"Position":     &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "Engineer"}}},
			"Description":  &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "Some desc"}}},
			"Status":       &notionapi.MultiSelectProperty{MultiSelect: []notionapi.Option{{Name: "Applied"}}},
			"Cover Letter": &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "Dear hiring manager"}}},
		},
	}
	jf, af := MapPage(page)
	if jf.Company != "Acme" {
		t.Errorf("Company: got %q, want %q", jf.Company, "Acme")
	}
	if jf.Position != "Engineer" {
		t.Errorf("Position: got %q, want %q", jf.Position, "Engineer")
	}
	if jf.Description != "Some desc" {
		t.Errorf("Description: got %q, want %q", jf.Description, "Some desc")
	}
	if af.Status != "Applied" {
		t.Errorf("Status: got %q, want %q", af.Status, "Applied")
	}
	if af.CoverLetter != "Dear hiring manager" {
		t.Errorf("CoverLetter: got %q, want %q", af.CoverLetter, "Dear hiring manager")
	}
}

func TestMapPage_CompanyAsTitleOrRichText(t *testing.T) {
	// Company supplied as RichText (Title extraction falls back to RichText)
	page := notionapi.Page{
		ID: "page-456",
		Properties: notionapi.Properties{
			"Company":  &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "FallbackCo"}}},
			"Position": &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "Dev"}}},
		},
	}
	jf, _ := MapPage(page)
	if jf.Company != "FallbackCo" {
		t.Errorf("Company fallback: got %q, want %q", jf.Company, "FallbackCo")
	}
}

func TestMapPage_PositionAsTitleOrRichText(t *testing.T) {
	// Position supplied as Title (RichText extraction falls back to Title)
	page := notionapi.Page{
		ID: "page-789",
		Properties: notionapi.Properties{
			"Company":  &notionapi.TitleProperty{Title: []notionapi.RichText{{PlainText: "Corp"}}},
			"Position": &notionapi.TitleProperty{Title: []notionapi.RichText{{PlainText: "Manager"}}},
		},
	}
	jf, _ := MapPage(page)
	if jf.Position != "Manager" {
		t.Errorf("Position fallback: got %q, want %q", jf.Position, "Manager")
	}
}

func TestMapPage_MissingFields(t *testing.T) {
	page := notionapi.Page{
		ID:         "page-empty",
		Properties: notionapi.Properties{},
	}
	jf, af := MapPage(page)
	if jf.Company != "" || jf.Position != "" || jf.Description != "" {
		t.Errorf("expected empty JobFields, got %+v", jf)
	}
	if af.Status != "" || af.CoverLetter != "" {
		t.Errorf("expected empty AppFields, got %+v", af)
	}
}

func TestMapPage_PageIDPropagated(t *testing.T) {
	page := notionapi.Page{ID: "abc-123", Properties: notionapi.Properties{}}
	_, af := MapPage(page)
	if af.NotionPageID != "abc-123" {
		t.Errorf("NotionPageID: got %q, want %q", af.NotionPageID, "abc-123")
	}
}
