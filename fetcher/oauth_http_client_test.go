package fetcher_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/pivotal-cf/p-mysql-manifest-validation/fetcher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewOAuthHTTPClient", func() {
	It("returns a client that can make authencated requests", func() {
		oauthServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/oauth/token"))
			username, password, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("opsman"))
			Expect(password).To(BeEmpty())

			err := req.ParseForm()
			Expect(err).NotTo(HaveOccurred())
			Expect(req.Form).To(Equal(url.Values{
				"client_id":  []string{"opsman"},
				"grant_type": []string{"password"},
				"username":   []string{"opsman-username"},
				"password":   []string{"opsman-password"},
			}))

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
		}))

		client, err := fetcher.NewOAuthHTTPClient(oauthServer.URL, "opsman-username", "opsman-password")
		Expect(err).NotTo(HaveOccurred())

		var wasCalled bool
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			wasCalled = true
			Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-opsman-token"))
		}))

		request, err := http.NewRequest("GET", server.URL, nil)
		Expect(err).NotTo(HaveOccurred())

		_, err = client.Do(request)
		Expect(err).NotTo(HaveOccurred())
		Expect(wasCalled).To(BeTrue())
	})

	Context("failure cases", func() {
		Context("when the token cannot be retrieved", func() {
			It("returns an error", func() {
				oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}))

				_, err := fetcher.NewOAuthHTTPClient(oauthServer.URL, "opsman-username", "opsman-password")
				Expect(err).To(MatchError(ContainSubstring("oauth2: cannot fetch token: 418 I'm a teapot")))
			})
		})
	})
})
