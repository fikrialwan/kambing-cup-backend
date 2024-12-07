package helper

import "strings"

func FormatSlug(slug string) string {
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}
