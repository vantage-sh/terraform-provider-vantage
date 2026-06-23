package vantage

import "testing"

func TestPageFromURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    int32
		wantErr bool
	}{
		{
			name:   "standard next-page URL",
			rawURL: "https://api.vantage.sh/v2/integrations?limit=1000&page=2",
			want:   2,
		},
		{
			name:   "page 10",
			rawURL: "https://api.vantage.sh/v2/integrations?page=10&limit=1000&provider=custom_provider",
			want:   10,
		},
		{
			name:    "missing page parameter",
			rawURL:  "https://api.vantage.sh/v2/integrations?limit=1000",
			wantErr: true,
		},
		{
			name:    "non-integer page value",
			rawURL:  "https://api.vantage.sh/v2/integrations?page=two",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			rawURL:  "://bad-url",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := pageFromURL(tc.rawURL)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%d)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got page %d, want %d", got, tc.want)
			}
		})
	}
}
