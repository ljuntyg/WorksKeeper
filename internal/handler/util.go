package handler

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Case 1 (localhost):
// request made to localhost -> Nginx proxies to localhost:8080 with X-Suffix-Subdomain = "" -> localhost:8080 redirects to localhost:8080/works

// Case 2 (localhost/resource):
// redirect request made to localhost/works -> Nginx proxies to localhost:8080 with X-Suffix-Subdomain = "works"
// -> Go subdomainPeriodReplacer redirects to works.at.localhost -> Nginx proxies to localhost:8080 with X-Prefix-Subdomain = "works.at"
// -> Go subdomainPeriodReplacer handles and responds to request to works.at.localhost by replacing the request url with localhost/works internally

// Case 3 (subdomain.localhost/resource):
// request made to works.at.localhost/work/2760 -> Nginx proxies to localhost:8080 with X-Prefix-Subdomain = "works.at" and X-Suffix-Subdomain = "work/2760"
// -> Go subdomainPeriodReplacer redirects to work.2760.at.localhost -> Nginx proxies to localhost:8080 with X-Prefix-Subdomain = "work.2760.at"
// -> Go subdomainPeriodReplacer handles and responds to request made to work.2760.at.localhost by replacing the request url with localhost/work/2760 internally
func SubdomainPeriodReplacer(scheme, host string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suffixSubdomain := r.Header.Get("X-Suffix-Subdomain")
		prefixSubdomain := r.Header.Get("X-Prefix-Subdomain")
		log.Printf("suffix: %s, prefix: %s", suffixSubdomain, prefixSubdomain)
		log.Println("path: " + r.URL.Path)

		if suffixSubdomain != "" {
			subdomain := strings.ReplaceAll(suffixSubdomain, "/", ".")
			target := url.URL{
				Scheme: scheme,
				Host:   subdomain + ".at." + host,
			}

			log.Println("Redirecting to URL: " + target.String())
			http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)
			return
		} else if prefixSubdomain != "" {
			r.URL.Path = "/" + strings.ReplaceAll(prefixSubdomain, ".", "/")[0:len(prefixSubdomain)-len(".at.")+1]
		}

		log.Println("Continuing to URL: " + r.Host + r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

// withoutContentSniffing binds a browser to the Content-Type a stored file is
// served under. The contents of an uploaded file are whatever the client sent,
// while its type is the one it claimed to be sending, so a browser left free to
// read the bytes and decide for itself could be talked into treating one of them
// as a document of our own origin.
func WithoutContentSniffing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(rw, r)
	})
}

func loadTemplate(filePath string) *template.Template {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return tmpl
}

func mustNumberingStringToId(numberingString string) int64 {
	if numberingString == "" {
		log.Panicln("no numbering in path")
	}

	idLen, err := strconv.Atoi(numberingString[:1])
	if err != nil {
		log.Panicln("invalid numbering")
	}

	id, idErr := strconv.Atoi(numberingString[1 : 1+idLen])
	if idErr != nil {
		log.Panicln("unable to parse id")
	}

	return int64(id)
}

// Assumes all inputs (which should be buttons) with name "action" are of the format
// "action-name: value"
func extractActionAndValueFromRequest(r *http.Request) (string, string) {
	actionString := r.FormValue("action")
	action, value, found := strings.Cut(actionString, ":")
	if !found {
		panic("unexpected action name when handling action")
	}

	return action, value
}

// extractOptionalFieldFromRequest tells a field that was left empty apart from
// one that was never rendered: a page shown in view mode carries no textareas
// at all, and a field that isn't there has to leave the stored value alone
// rather than blank it. Assumes the form has already been parsed.
func extractOptionalFieldFromRequest(r *http.Request, name string) *string {
	values, found := r.PostForm[name]
	if !found || len(values) == 0 {
		return nil
	}

	return &values[0]
}

// extractIndexedFieldsFromRequest collects the fields that carry the contents
// of a row, which are named "<name>[<row id>]" so that every row on the page
// submits under its own name. Assumes the form has already been parsed.
func extractIndexedFieldsFromRequest(r *http.Request, name string) map[int64]string {
	fields := make(map[int64]string)
	prefix := name + "["

	for fieldName, values := range r.PostForm {
		if !strings.HasPrefix(fieldName, prefix) || !strings.HasSuffix(fieldName, "]") || len(values) == 0 {
			continue
		}

		id, err := strconv.ParseInt(fieldName[len(prefix):len(fieldName)-1], 10, 64)
		if err != nil {
			panic("unexpected id in form field name")
		}

		fields[id] = values[0]
	}

	return fields
}
