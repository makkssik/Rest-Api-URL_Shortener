package tests

import (
	"RestApi/internal/http-server/handlers/url/save"
	"RestApi/internal/lib/api"
	"RestApi/internal/lib/random"
	"net/http"
	"net/url"
	"path"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/url").WithJSON(save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}).WithBasicAuth("myuser", "mypass").
		Expect().
		Status(200).
		JSON().
		Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is invalid url",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			req := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("myuser", "mypass").
				Expect().Status(http.StatusOK).JSON().Object()

			if tc.error != "" {
				req.NotContainsKey("alias")

				req.Value("error").String().IsEqual(tc.error)
			}

			if tc.error != "" {
				return
			}

			alias := tc.alias

			if tc.alias != "" {
				req.Value("alias").String().IsEqual(tc.alias)
			} else {
				req.Value("alias").String().NotEmpty()

				alias = req.Value("alias").String().Raw()
			}

			if tc.error != "" {
				return
			}

			testRedirect(t, alias, tc.url)

			reqDel := e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("myuser", "mypass").
				Expect().Status(http.StatusOK).
				JSON().Object()

			reqDel.Value("status").String().IsEqual("OK")

			testRedirectNotFound(t, alias)
		})
	}
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirect(u.String())
	require.Error(t, err)
}

func testRedirect(t *testing.T, alias string, expectedURL string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, expectedURL, redirectedToURL)
}
