package core

import (
	"fmt"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var nameStripRE = regexp.MustCompile("(?i)^((20)|(25)|(2b)|(2f)|(3d)|(3a)|(40))+")

func GetRawCookie(cookies []*http.Cookie) string {
	var rawCookies []string
	for _, c := range cookies {
		e := fmt.Sprintf("%s=%s", c.Name, c.Value)
		rawCookies = append(rawCookies, e)
	}
	return strings.Join(rawCookies, "; ")
}

func GetDomain(site *url.URL) string {
	domain, err := publicsuffix.EffectiveTLDPlusOne(site.Hostname())
	if err != nil {
		return ""
	}
	return domain
}

func FixUrl(url string, site *url.URL) string {
	var newUrl string
	if strings.HasPrefix(url, "//") {
		// //google.com/example.php
		newUrl = site.Scheme + ":" + url

	} else if strings.HasPrefix(url, "http") {
		// http://google.com || https://google.com
		newUrl = url

	} else if !strings.HasPrefix(url, "//") {
		if strings.HasPrefix(url, "/") {
			// Ex: /?thread=10
			newUrl = site.Scheme + "://" + site.Host + url

		} else {
			if strings.HasPrefix(url, ".") {
				if strings.HasPrefix(url, "..") {
					newUrl = site.Scheme + "://" + site.Host + url[2:]
				} else {
					newUrl = site.Scheme + "://" + site.Host + url[1:]
				}
			} else {
				// "console/test.php"
				newUrl = site.Scheme + "://" + site.Host + "/" + url
			}
		}
	}
	return newUrl
}

func Unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func LoadCookies(rawCookie string) []*http.Cookie {
	httpCookies := []*http.Cookie{}
	cookies := strings.Split(rawCookie, ";")
	for _, cookie := range cookies {
		cookieArgs := strings.SplitN(cookie, "=", 2)
		if len(cookieArgs) > 2 {
			continue
		}

		ck := &http.Cookie{Name: strings.TrimSpace(cookieArgs[0]), Value: strings.TrimSpace(cookieArgs[1])}
		httpCookies = append(httpCookies, ck)
	}
	return httpCookies
}

func GetExtType(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return ""
	}
	return path.Ext(u.Path)
}

func CleanSubdomain(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimPrefix(s, "*.")
	//s = strings.Trim("u00","")
	s = cleanName(s)
	return s
}

// Clean up the names scraped from the web.
// Get from Amass
func cleanName(name string) string {
	for {
		if i := nameStripRE.FindStringIndex(name); i != nil {
			name = name[i[1]:]
		} else {
			break
		}
	}

	name = strings.Trim(name, "-")
	// Remove dots at the beginning of names
	if len(name) > 1 && name[0] == '.' {
		name = name[1:]
	}
	return name
}

func FilterNewLines(s string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), " ")
}

func DecodeChars(s string) string {
	source, err := url.QueryUnescape(s)
	if err == nil {
		s = source
	}

	// In case json encoded chars
	replacer := strings.NewReplacer(
		`\u002f`, "/",
		`\U002F`, "/",
		`\u002F`, "/",
		`\u0026`, "&",
		`\U0026`, "&",
	)
	s = replacer.Replace(s)
	return s
}

func InScope(u *url.URL, regexps []*regexp.Regexp) bool {
	for _, r := range regexps {
		if r.MatchString(u.Hostname()) {
			return true
		}
	}
	return false
}
