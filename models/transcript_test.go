package models

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		// Valid URLs - YouTube
		{"YouTube standard", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"YouTube short", "https://youtu.be/dQw4w9WgXcQ", true},
		{"YouTube no www", "https://youtube.com/watch?v=dQw4w9WgXcQ", true},

		// Valid URLs - Other platforms supported by yt-dlp
		{"Vimeo", "https://vimeo.com/123456789", true},
		{"Dailymotion", "https://www.dailymotion.com/video/x123abc", true},
		{"Twitter", "https://twitter.com/user/status/123456789", true},
		{"TikTok", "https://www.tiktok.com/@user/video/123456789", true},
		{"Twitch VOD", "https://www.twitch.tv/videos/123456789", true},
		{"SoundCloud", "https://soundcloud.com/artist/track", true},
		{"Instagram", "https://www.instagram.com/p/ABC123/", true},

		// Valid URLs - HTTP (not just HTTPS)
		{"HTTP URL", "http://example.com/video", true},

		// Invalid URLs
		{"Empty string", "", false},
		{"Not a URL", "not-a-url", false},
		{"FTP scheme", "ftp://example.com/file", false},
		{"File scheme", "file:///path/to/file", false},
		{"No scheme", "www.youtube.com/watch?v=abc", false},
		{"No host", "https:///path", false},
		{"Just scheme", "https://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateURL(tt.url)
			if result != tt.expected {
				t.Errorf("ValidateURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}
