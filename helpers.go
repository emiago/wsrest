package wsrest

import (
	"fmt"
	"net/url"
)

func UrlComponentEncode(uri string, args ...interface{}) string {
	escaped := make([]interface{}, len(args))
	for k, v := range args {
		switch vt := v.(type) {
		case string:
			escaped[k] = url.PathEscape(vt)
		default:
			escaped[k] = v
		}
	}

	return fmt.Sprintf(uri, escaped...)
}

func copyStringMap(source map[string]string, target map[string]string) map[string]string {
	if target == nil {
		target = make(map[string]string)
	}

	for k, v := range source {
		target[k] = v
	}
	return target
}
