package utils

import "net/url"

func ExtractDomain(rawurl string) (string, error) {
	parsedUrl, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	return parsedUrl.Host, nil
}
