package discord

import (
	"testing"

	"job-tracker/internal/entity"
)

func TestParseAdd(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want addArgs
		ok   bool
	}{
		{"three fields", "Backend Engineer | Tokopedia | build stuff", addArgs{"Backend Engineer", "Tokopedia", "build stuff", ""}, true},
		{"four fields", "Backend Engineer | Tokopedia | build stuff | Interview", addArgs{"Backend Engineer", "Tokopedia", "build stuff", "Interview"}, true},
		{"whitespace trimmed", "  BE  |  Toko  |  desc  ", addArgs{"BE", "Toko", "desc", ""}, true},
		{"too few fields", "BE | Toko", addArgs{}, false},
		{"empty field", "BE |  | desc", addArgs{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseAdd(tt.in)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

type stubStatusRepo struct{ statuses []entity.Status }

func (s stubStatusRepo) FindAll() ([]entity.Status, error) { return s.statuses, nil }

func TestResolveStatusID(t *testing.T) {
	b := &Bot{statusRepo: stubStatusRepo{statuses: []entity.Status{
		{ID: 1, Text: "Applied"}, {ID: 3, Text: "Interview"},
	}}}
	tests := []struct {
		name string
		want int
	}{
		{"", 1},              // empty → default Applied
		{"Interview", 3},     // known → its id
		{"interview", 3},     // case-insensitive
		{"Nonexistent", 1},   // unknown → default
	}
	for _, tt := range tests {
		got, err := b.resolveStatusID(tt.name)
		if err != nil {
			t.Fatalf("resolveStatusID(%q) error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("resolveStatusID(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}
