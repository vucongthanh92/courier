package utils

import (
	"net/url"
)

func SanitizedOAuthRedirectURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsedURL.RawFragment == "" {
		return parsedURL.String()
	}

	fragment := url.Values{}
	for key, values := range ParseFragmentValues(parsedURL.RawFragment) {
		for _, value := range values {
			fragment.Add(key, value)
		}
	}

	MaskQueryValue(fragment, "oauth_result")
	parsedURL.RawFragment = fragment.Encode()
	return parsedURL.String()
}

func ParseFragmentValues(fragment string) url.Values {
	values, err := url.ParseQuery(fragment)
	if err != nil {
		return url.Values{"fragment": []string{"<unparseable>"}}
	}
	return values
}

func MaskQueryValue(values url.Values, key string) {
	if values.Get(key) == "" {
		return
	}
	values.Set(key, "<redacted>")
}
