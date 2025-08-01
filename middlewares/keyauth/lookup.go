package keyauth

import (
	"errors"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/savsgio/gotils/strconv"
	"strings"
)

type LookupToken func(*app.RequestContext) (string, error)

// LookupTokenFunc return function LookupToken
func LookupTokenFunc(option *Options) LookupToken {
	parts := strings.Split(option.keyLookup, ":")
	if len(parts) != 2 {
		panic(errors.New("the length of parts should be equal to 2"))
	}
	extractor := KeyFromHeader(parts[1], option.authScheme)
	switch parts[0] {
	case "query":
		extractor = KeyFromQuery(parts[1])
	case "form":
		extractor = KeyFromForm(parts[1])
	case "param":
		extractor = KeyFromParam(parts[1])
	case "cookie":
		extractor = KeyFromCookie(parts[1])
	}
	return extractor
}

// KeyFromHeader returns a function that extracts api key from the request header.
func KeyFromHeader(header, authScheme string) LookupToken {
	return func(c *app.RequestContext) (string, error) {
		auth := strconv.B2S(c.GetHeader(header))
		l := len(authScheme)
		if len(auth) > 0 && l == 0 {
			return auth, nil
		}
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", ErrMissingOrMalformedAPIKey
	}
}

// KeyFromQuery returns a function that extracts api key from the query string.
func KeyFromQuery(param string) LookupToken {
	return func(c *app.RequestContext) (string, error) {
		key := c.Query(param)
		if key == "" {
			return "", ErrMissingOrMalformedAPIKey
		}
		return key, nil
	}
}

// KeyFromForm returns a function that extracts api key from the form.
func KeyFromForm(param string) LookupToken {
	return func(c *app.RequestContext) (string, error) {
		key := strconv.B2S(c.FormValue(param))
		if key == "" {
			return "", ErrMissingOrMalformedAPIKey
		}
		return key, nil
	}
}

// KeyFromParam returns a function that extracts api key from the url param string.
func KeyFromParam(param string) LookupToken {
	return func(c *app.RequestContext) (string, error) {
		key := c.Param(param)
		if key == "" {
			return "", ErrMissingOrMalformedAPIKey
		}
		return key, nil
	}
}

// KeyFromCookie returns a function that extracts api key from the named cookie.
func KeyFromCookie(name string) LookupToken {
	return func(c *app.RequestContext) (string, error) {
		key := strconv.B2S(c.Cookie(name))
		if key == "" {
			return "", ErrMissingOrMalformedAPIKey
		}
		return key, nil
	}
}
