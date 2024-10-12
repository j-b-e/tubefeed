package utils

import "testing"

func TestExtractDomain(t *testing.T) {
	cases := map[string]string{
		"http://youtube.com?asdasd=asdas": "youtube.com",
		"http://www.youtube.com?asdasdas": "youtube.com",
		"https://m.youtube.com?asdasdas":  "youtube.com",
	}

	for k, v := range cases {
		domain, err := ExtractDomain(k)
		if err != nil {
			t.Fatalf("ExtractDomain Error: %v", err)
		}
		if domain != v {
			t.Errorf("ExtractDomain does not match: %s != %s", domain, v)
		}
	}
}
