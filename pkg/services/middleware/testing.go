package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/macaron.v1"

	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/auth"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/require"
)

type scenarioContext struct {
	service              *MiddlewareService
	m                    *macaron.Macaron
	context              *models.ReqContext
	resp                 *httptest.ResponseRecorder
	apiKey               string
	authHeader           string
	tokenSessionCookie   string
	respJson             map[string]interface{}
	handlerFunc          handlerFunc
	defaultHandler       macaron.Handler
	url                  string
	userAuthTokenService *auth.FakeUserAuthTokenService
	remoteCacheService   *remotecache.RemoteCache

	req *http.Request
}

func (sc *scenarioContext) withValidApiKey() *scenarioContext {
	sc.apiKey = "eyJrIjoidjVuQXdwTWFmRlA2em5hUzR1cmhkV0RMUzU1MTFNNDIiLCJuIjoiYXNkIiwiaWQiOjF9"
	return sc
}

func (sc *scenarioContext) withTokenSessionCookie(unhashedToken string) *scenarioContext {
	sc.tokenSessionCookie = unhashedToken
	return sc
}

func (sc *scenarioContext) withAuthorizationHeader(authHeader string) *scenarioContext {
	sc.authHeader = authHeader
	return sc
}

func (sc *scenarioContext) fakeReq(t *testing.T, method, url string) *scenarioContext {
	sc.resp = httptest.NewRecorder()
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)
	sc.req = req

	return sc
}

func (sc *scenarioContext) fakeReqWithParams(t *testing.T, method, url string, queryParams map[string]string) *scenarioContext {
	sc.resp = httptest.NewRecorder()
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)

	q := req.URL.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	sc.req = req

	return sc
}

func (sc *scenarioContext) handler(fn handlerFunc) *scenarioContext {
	sc.handlerFunc = fn
	return sc
}

func (sc *scenarioContext) exec(t *testing.T) {
	if sc.apiKey != "" {
		sc.req.Header.Add("Authorization", "Bearer "+sc.apiKey)
	}

	if sc.authHeader != "" {
		sc.req.Header.Add("Authorization", sc.authHeader)
	}

	if sc.tokenSessionCookie != "" {
		sc.req.AddCookie(&http.Cookie{
			Name:  setting.LoginCookieName,
			Value: sc.tokenSessionCookie,
		})
	}

	sc.m.ServeHTTP(sc.resp, sc.req)

	if sc.resp.Header().Get("Content-Type") == "application/json; charset=UTF-8" {
		err := json.NewDecoder(sc.resp.Body).Decode(&sc.respJson)
		require.NoError(t, err)
	}
}

type handlerFunc func(c *models.ReqContext)