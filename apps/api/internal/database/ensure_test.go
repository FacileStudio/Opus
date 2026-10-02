package database

import "testing"

func TestSplitDatabaseURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantAdmin string
		wantDB    string
		wantOK    bool
	}{
		{
			name:      "postgresql scheme",
			url:       "postgresql://opus:secret@postgres:5432/opus_go",
			wantAdmin: "postgresql://opus:secret@postgres:5432/postgres",
			wantDB:    "opus_go",
			wantOK:    true,
		},
		{
			name:      "postgres scheme keeps query",
			url:       "postgres://opus:secret@postgres:5432/opus_go?sslmode=disable",
			wantAdmin: "postgres://opus:secret@postgres:5432/postgres?sslmode=disable",
			wantDB:    "opus_go",
			wantOK:    true,
		},
		{
			name:   "no database name",
			url:    "postgresql://opus:secret@postgres:5432",
			wantOK: false,
		},
		{
			name:   "empty database name",
			url:    "postgresql://opus:secret@postgres:5432/",
			wantOK: false,
		},
		{
			name:   "keyword dsn is left alone",
			url:    "host=postgres user=opus dbname=opus_go",
			wantOK: false,
		},
		{
			name:   "non postgres url",
			url:    "mysql://opus:secret@postgres:3306/opus_go",
			wantOK: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			admin, name, ok := splitDatabaseURL(test.url)
			if ok != test.wantOK || admin != test.wantAdmin || name != test.wantDB {
				t.Fatalf("splitDatabaseURL(%q) = (%q, %q, %v), want (%q, %q, %v)",
					test.url, admin, name, ok, test.wantAdmin, test.wantDB, test.wantOK)
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	if got := quoteIdentifier("opus_go"); got != `"opus_go"` {
		t.Fatalf("quoteIdentifier(opus_go) = %q", got)
	}
	if got := quoteIdentifier(`we"ird`); got != `"we""ird"` {
		t.Fatalf("quoteIdentifier with a quote = %q", got)
	}
}
