package router

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"imgcutter/service"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
)

type requestCheckParamsLog struct {
	hasCtx      bool
	ctxValue    string
	hasCookie   bool
	cookieValue string
}

func getMockAssertFunctionLogging(t *testing.T, reqParams requestCheckParamsLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxValue, ok := r.Context().Value(ctxSessionKey).(string)
		if reqParams.hasCtx {
			assert.Equal(t, ok, true)
			assert.Equal(t, ctxValue, reqParams.ctxValue)
		} else {
			assert.Equal(t, ok, false)
		}

		c, err := r.Cookie(sessionID)
		if reqParams.hasCookie {
			assert.Equal(t, err, nil)
			assert.Equal(t, c.Value, reqParams.cookieValue)
		} else {
			assert.Equal(t, err, http.ErrNoCookie)
		}
	}
}

func TestRouter_MiddelwareLogging(t *testing.T) {
	testCases := []struct {
		name         string
		host         string
		url          string
		method       string
		params       requestCheckParamsLog
		logOutRegexp string
	}{
		{
			name:   "simple req",
			host:   "example.com",
			url:    "/",
			method: http.MethodGet,
			params: requestCheckParamsLog{
				hasCtx:      false,
				ctxValue:    "",
				hasCookie:   false,
				cookieValue: "",
			},
			logOutRegexp: "request: example.com GET /",
		},
		{
			name:   "req with ctx value",
			host:   "ru.example.org",
			url:    "/",
			method: http.MethodGet,
			params: requestCheckParamsLog{
				hasCtx:      true,
				ctxValue:    "value",
				hasCookie:   false,
				cookieValue: "",
			},
			logOutRegexp: "request: ru.example.org GET /",
		},
		{
			name:   "cookie",
			host:   "sait.ru",
			url:    "/",
			method: http.MethodGet,
			params: requestCheckParamsLog{
				hasCtx:      false,
				ctxValue:    "",
				hasCookie:   true,
				cookieValue: "value",
			},
			logOutRegexp: "request: sait.ru GET /",
		},
		{
			name:   "another request",
			host:   "localhost:8080",
			url:    "/some/path",
			method: http.MethodPost,
			params: requestCheckParamsLog{
				hasCtx:      true,
				ctxValue:    "ctxValue",
				hasCookie:   true,
				cookieValue: "cookieValue",
			},
			logOutRegexp: "request: localhost:8080 POST /some/path",
		},
		{
			name:   "another request",
			url:    "/abra-cadabra",
			method: http.MethodPut,
			params: requestCheckParamsLog{
				hasCtx:      false,
				ctxValue:    "",
				hasCookie:   false,
				cookieValue: "",
			},
			logOutRegexp: "request:  PUT /abra-cadabra",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			buf := bytes.Buffer{}
			log.SetOutput(&buf)

			c := gomock.NewController(t)
			defer c.Finish()

			ss := service.NewMockSessionService(c)
			fs := service.NewMockFileService(c)
			te := NewMocktemplateExecutor(c)
			handler := Handler{
				templates: te,
				service:   service.Service{Files: fs, Session: ss},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(tc.method, tc.url, nil)
			r.Host = tc.host

			assertHandleFunc := getMockAssertFunctionLogging(t, tc.params)
			composition := handler.Logging(assertHandleFunc)

			if tc.params.hasCtx {
				r = r.WithContext(context.WithValue(context.Background(), ctxSessionKey, tc.params.ctxValue))
			}

			if tc.params.hasCookie {
				r.AddCookie(&http.Cookie{Name: sessionID, Value: tc.params.cookieValue})
			}

			composition(w, r)

			assert.Matches(t, buf.String(), tc.logOutRegexp)
		})
	}
}

func getMockAssertFunctionSession(t *testing.T, wantedCtxValue string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxValue, ok := r.Context().Value(ctxSessionKey).(string)
		assert.Equal(t, ok, true)
		assert.Equal(t, ctxValue, wantedCtxValue)
	}
}

func TestRouter_MiddelwareManageSession(t *testing.T) {
	testCases := []struct {
		name                    string
		wantedCtxValue          string
		shouldSetCookies        bool
		wantedSetCookieValue    string
		serviceSessionBehaviour func(mss *service.MockSessionService)
		requestSetup            func(r *http.Request)
		logOutRegexp            string
	}{
		{
			name:                 "no cookie request",
			wantedCtxValue:       "00000000-0000-0000-0000-000000000000",
			shouldSetCookies:     true,
			wantedSetCookieValue: "00000000-0000-0000-0000-000000000000",
			serviceSessionBehaviour: func(mss *service.MockSessionService) {
				mss.EXPECT().New().Return(&service.Session{})
			},
			requestSetup: func(r *http.Request) {
			},
			logOutRegexp: "session cookie not found\n.+creating new session\n.+sent cookie: SESSID=00000000-0000-0000-0000-000000000000; Path=/; HttpOnly\n.+working session: 00000000-0000-0000-0000-000000000000",
		},
		{
			name:                 "request with valid cookie",
			wantedCtxValue:       "01234567-ffff-0123-0123-0123456789abc",
			shouldSetCookies:     false,
			wantedSetCookieValue: "",
			serviceSessionBehaviour: func(mss *service.MockSessionService) {
				mss.EXPECT().Find("01234567-ffff-0123-0123-0123456789abc").Return(&service.Session{}, true)
			},
			requestSetup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: sessionID, Value: "01234567-ffff-0123-0123-0123456789abc"})
			},
			logOutRegexp: "working session: 01234567-ffff-0123-0123-0123456789abc",
		},
		{
			name:                 "invalid session cookie request",
			wantedCtxValue:       "00000000-0000-0000-0000-000000000000",
			shouldSetCookies:     true,
			wantedSetCookieValue: "00000000-0000-0000-0000-000000000000",
			serviceSessionBehaviour: func(mss *service.MockSessionService) {
				mss.EXPECT().Find("00000000-ffff-1111-eeee-555566667777").Return(&service.Session{}, false)
				mss.EXPECT().New().Return(&service.Session{})
			},
			requestSetup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: sessionID, Value: "00000000-ffff-1111-eeee-555566667777"})
			},
			logOutRegexp: "creating new session\n.+sent cookie: SESSID=00000000-0000-0000-0000-000000000000; Path=/; HttpOnly\n.+working session: 00000000-0000-0000-0000-000000000000",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			buf := bytes.Buffer{}
			log.SetOutput(&buf)

			c := gomock.NewController(t)
			defer c.Finish()

			ss := service.NewMockSessionService(c)
			fs := service.NewMockFileService(c)
			te := NewMocktemplateExecutor(c)
			handler := Handler{
				templates: te,
				service:   service.Service{Files: fs, Session: ss},
			}
			tc.serviceSessionBehaviour(ss)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			tc.requestSetup(r)

			assertHandleFunc := getMockAssertFunctionSession(t, tc.wantedCtxValue)
			composition := handler.ManageSession(assertHandleFunc)

			composition(w, r)

			assert.Matches(t, buf.String(), tc.logOutRegexp)

			outCookies := w.Result().Cookies()
			if tc.shouldSetCookies {
				assert.Equal(t, len(outCookies), 1)
				assert.Equal(t, outCookies[0].Value, tc.wantedSetCookieValue)
			} else {
				assert.Equal(t, len(outCookies), 0)
			}
		})
	}
}
