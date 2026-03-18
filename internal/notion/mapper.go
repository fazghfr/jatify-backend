package notion

import (
	"github.com/jomei/notionapi"
)

type JobFields struct {
	Company     string
	Position    string
	Description string
}

type AppFields struct {
	NotionPageID string
	Status       string
	CoverLetter  string
}

func ExtractTitle(prop notionapi.Property) string {
	if t, ok := prop.(*notionapi.TitleProperty); ok {
		if len(t.Title) > 0 {
			return t.Title[0].PlainText
		}
	}
	return ""
}

func ExtractRichText(prop notionapi.Property) string {
	if rt, ok := prop.(*notionapi.RichTextProperty); ok {
		if len(rt.RichText) > 0 {
			return rt.RichText[0].PlainText
		}
	}
	return ""
}

func ExtractSelect(prop notionapi.Property) string {
	if s, ok := prop.(*notionapi.SelectProperty); ok && s.Select.Name != "" {
		return s.Select.Name
	}
	return ""
}

// ExtractMultiSelectLast returns the name of the last option in a MultiSelectProperty,
// or "" if the property is the wrong type or the list is empty.
func ExtractMultiSelectLast(prop notionapi.Property) string {
	if ms, ok := prop.(*notionapi.MultiSelectProperty); ok && len(ms.MultiSelect) > 0 {
		return ms.MultiSelect[len(ms.MultiSelect)-1].Name
	}
	return ""
}

// MapPage extracts JobFields and AppFields from a Notion page.
// Returns empty JobFields (Company=="") if required fields are missing.
func MapPage(page notionapi.Page) (JobFields, AppFields) {
	var jf JobFields
	var af AppFields

	af.NotionPageID = string(page.ID)

	if prop, ok := page.Properties["Company"]; ok {
		jf.Company = ExtractTitle(prop)
		if jf.Company == "" {
			jf.Company = ExtractRichText(prop)
		}
	}
	if prop, ok := page.Properties["Position"]; ok {
		jf.Position = ExtractRichText(prop)
		if jf.Position == "" {
			jf.Position = ExtractTitle(prop)
		}
	}
	if prop, ok := page.Properties["Description"]; ok {
		jf.Description = ExtractRichText(prop)
	}
	if prop, ok := page.Properties["Status"]; ok {
		af.Status = ExtractMultiSelectLast(prop)
	}
	if prop, ok := page.Properties["Cover Letter"]; ok {
		af.CoverLetter = ExtractRichText(prop)
	}

	return jf, af
}
