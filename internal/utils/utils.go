package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func ExtractDomain(rawurl string) (string, error) {
	parsedUrl, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	host := strings.Split(parsedUrl.Host, ".")
	if len(host) < 2 {
		return "", fmt.Errorf("%s ist not a fqdn", parsedUrl.Host)
	}
	return strings.Join(host[len(host)-2:], "."), nil
}

func StringToPointer(s string) *string {
	return &s
}
