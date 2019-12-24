package util

import (
	"github.com/thoas/go-funk"
	"net/url"
)

// NormalizeURL get an url and returns only scheme://host
func NormalizeURL(s []interface{}) []interface{} {
	return funk.Map(s, func(s interface{}) interface{} {
		u, err := url.Parse(s.(string))
		if err != nil {
			return s
		}
		var n url.URL
		n.Scheme = u.Scheme
		n.Host = u.Host

		return n.String()
	}).([]interface{})
}
