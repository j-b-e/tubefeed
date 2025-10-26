package yt

import "testing"

func Test_extractVideoID(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "standard URL",
			url:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "shortened URL",
			url:     "https://youtu.be/dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: true, // TODO support this case
		},
		{
			name:    "embed URL",
			url:     "https://www.youtube.com/embed/dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: true, // TODO support this case
		},
		{
			name:    "domain is ignored, parameters correct",
			url:     "https://www.example.com/watch?v=dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "no video ID",
			url:     "https://www.youtube.com/watch",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "additional parameters",
			url:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=42s",
			want:    "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "additional parameters order changed",
			url:     "https://www.youtube.com/watch?t=42s&v=dQw4w9WgXcQ&ab_channel=RickAstley",
			want:    "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "mobile URL",
			url:     "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
			want:    "dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "list parameter in URL",
			url:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PL1234567890",
			want:    "dQw4w9WgXcQ",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := extractVideoID(tt.url)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("extractVideoID() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("extractVideoID() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("extractVideoID() = %v, want %v", got, tt.want)
			}
		})
	}
}
