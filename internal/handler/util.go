package handler

import (
	"WorksKeeper/internal/mock"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var homeTemplate *template.Template

func init() {
	homeTemplate = loadTemplate("resources/home.html")
}

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
func SubdomainPeriodReplacer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suffixSubdomain := r.Header.Get("X-Suffix-Subdomain")
		prefixSubdomain := r.Header.Get("X-Prefix-Subdomain")
		log.Printf("suffix: %s, prefix: %s", suffixSubdomain, prefixSubdomain)
		log.Println("path: " + r.URL.Path)

		if suffixSubdomain != "" {
			subdomain := strings.ReplaceAll(suffixSubdomain, "/", ".")
			target := getProtocol() + subdomain + ".at." + getHost()
			log.Println("Redirecting to URL: " + target)
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		} else if prefixSubdomain != "" {
			r.URL.Path = "/" + strings.ReplaceAll(prefixSubdomain, ".", "/")[0:len(prefixSubdomain)-len(".at.")+1]
		}

		log.Println("Continuing to URL: " + r.Host + r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

// TODO:
func getProtocol() string {
	return mock.GetMockProtocol()
}

// TODO:
func getHost() string {
	return mock.GetMockHost()
}

func loadTemplate(filePath string) *template.Template {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return tmpl
}
