package router

import (
	"bytes"
	"context"
	"errors"
	"imgcutter/service"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
)

func TestRouter_CutFile(t *testing.T) {

	type cutParams struct {
		filename string
		dX       int
		dY       int
	}

	testCases := []struct {
		name                    string
		sessionID               string
		formContent             map[string]string
		cutParams               cutParams
		requestModification     func(r *http.Request, sessionID string) *http.Request
		sessionServiceBehaviour func(ss *service.MockSessionService, sessionID string)
		fileServiceBehaviour    func(fs *service.MockFileService, session *service.Session, cutParams cutParams)
		templateBehavior        func(te *MocktemplateExecutor, filename string)
		responseCode            int
	}{
		{
			name:        "ok",
			sessionID:   "random-uuid",
			formContent: map[string]string{"fileName": "filename", "dX": "250", "dY": "250"},
			cutParams:   cutParams{"filename", 250, 250},
			requestModification: func(r *http.Request, sessionID string) *http.Request {
				return r.WithContext(context.WithValue(context.Background(), ctxSessionKey, sessionID))
			},
			sessionServiceBehaviour: func(ss *service.MockSessionService, sessionID string) {
				ss.EXPECT().Find(sessionID).Return(&service.Session{}, true)
			},
			fileServiceBehaviour: func(fs *service.MockFileService, session *service.Session, cutParams cutParams) {
				fs.EXPECT().CutFile(session, cutParams.filename, cutParams.dX, cutParams.dY).Return(nil)
			},
			templateBehavior: func(te *MocktemplateExecutor, fileName string) {
				te.EXPECT().ExecuteTemplate(&bytes.Buffer{}, "cutGood.html", fileName).Return(nil)
			},
			responseCode: http.StatusOK,
		},
		{
			name:        "err parsing int",
			sessionID:   "random-uuid",
			formContent: map[string]string{"fileName": "filename", "dX": "dvesti", "dY": "250"},
			cutParams:   cutParams{},
			requestModification: func(r *http.Request, sessionID string) *http.Request {
				return r.WithContext(context.WithValue(context.Background(), ctxSessionKey, sessionID))
			},
			sessionServiceBehaviour: func(ss *service.MockSessionService, sessionID string) {
			},
			fileServiceBehaviour: func(fs *service.MockFileService, session *service.Session, cutParams cutParams) {
			},
			templateBehavior: func(te *MocktemplateExecutor, fileName string) {
			},
			responseCode: http.StatusBadRequest,
		},
		{
			name:        "no ctx value",
			sessionID:   "",
			formContent: map[string]string{"fileName": "filename", "dX": "250", "dY": "250"},
			cutParams:   cutParams{},
			requestModification: func(r *http.Request, sessionID string) *http.Request {
				return r
			},
			sessionServiceBehaviour: func(ss *service.MockSessionService, sessionID string) {
			},
			fileServiceBehaviour: func(fs *service.MockFileService, session *service.Session, cutParams cutParams) {
			},
			templateBehavior: func(te *MocktemplateExecutor, fileName string) {
			},
			responseCode: http.StatusInternalServerError,
		},
		{
			name:        "Session not found",
			sessionID:   "unknown-uuid",
			formContent: map[string]string{"fileName": "filename", "dX": "250", "dY": "250"},
			cutParams:   cutParams{},
			requestModification: func(r *http.Request, sessionID string) *http.Request {
				return r.WithContext(context.WithValue(context.Background(), ctxSessionKey, sessionID))
			},
			sessionServiceBehaviour: func(ss *service.MockSessionService, sessionID string) {
				ss.EXPECT().Find(sessionID).Return(&service.Session{}, false)
			},
			fileServiceBehaviour: func(fs *service.MockFileService, session *service.Session, cutParams cutParams) {
			},
			templateBehavior: func(te *MocktemplateExecutor, fileName string) {
			},
			responseCode: http.StatusNotFound,
		},
		{
			name:        "template fail",
			sessionID:   "random-uuid",
			formContent: map[string]string{"fileName": "filename", "dX": "250", "dY": "250"},
			cutParams:   cutParams{"filename", 250, 250},
			requestModification: func(r *http.Request, sessionID string) *http.Request {
				return r.WithContext(context.WithValue(context.Background(), ctxSessionKey, sessionID))
			},
			sessionServiceBehaviour: func(ss *service.MockSessionService, sessionID string) {
				ss.EXPECT().Find(sessionID).Return(&service.Session{}, true)
			},
			fileServiceBehaviour: func(fs *service.MockFileService, session *service.Session, cutParams cutParams) {
				fs.EXPECT().CutFile(session, cutParams.filename, cutParams.dX, cutParams.dY).Return(nil)
			},
			templateBehavior: func(te *MocktemplateExecutor, fileName string) {
				te.EXPECT().ExecuteTemplate(&bytes.Buffer{}, "cutGood.html", fileName).Return(errors.New("some err"))
			},
			responseCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			ss := service.NewMockSessionService(c)
			fs := service.NewMockFileService(c)
			te := NewMocktemplateExecutor(c)
			handler := Handler{
				templates: te,
				service:   service.Service{Files: fs, Session: ss},
			}

			tc.sessionServiceBehaviour(ss, tc.sessionID)
			tc.fileServiceBehaviour(fs, &service.Session{}, tc.cutParams)
			tc.templateBehavior(te, tc.cutParams.filename)

			params := url.Values{}
			for k, v := range tc.formContent {
				params.Add(k, v)
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/cut", bytes.NewBufferString(params.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			ctxr := tc.requestModification(r, tc.sessionID)

			handler.CutFile(w, ctxr)

			assert.Equal(t, w.Result().StatusCode, tc.responseCode)
		})
	}
}

func TestRouter_MainPage(t *testing.T) {
	testCases := []struct {
		name                    string
		sessionID               string
		requestModification     func(r *http.Request, sessionID string) *http.Request
		sessionServiceBehaviour func(mss *service.MockSessionService, sessionID string)
		fileServiceBehaviour    func(mfs *service.MockFileService, session *service.Session)
		templateBehavior        func(te *MocktemplateExecutor)
		responseCode            int
	}{
		{
			name:      "ok",
			sessionID: "random-uuid",
			requestModification: func(r *http.Request, sessionID string) *http.Request {
				return r.WithContext(context.WithValue(context.Background(), ctxSessionKey, sessionID))
			},
			sessionServiceBehaviour: func(mss *service.MockSessionService, sessionID string) {
				mss.EXPECT().Find(sessionID).Return(&service.Session{}, true)
			},
			fileServiceBehaviour: func(mfs *service.MockFileService, session *service.Session) {
				mfs.EXPECT().GetFiles(&service.Session{}).Return([]service.MyFile{{
					OriginalFile: "orig.jpg",
					Archive:      "orig.zip",
				}}, nil)
			},
			templateBehavior: func(te *MocktemplateExecutor) {
				te.EXPECT().ExecuteTemplate(&bytes.Buffer{}, "home.html", []service.MyFile{{
					OriginalFile: "orig.jpg",
					Archive:      "orig.zip",
				}}).Return(nil)
			},
			responseCode: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			ss := service.NewMockSessionService(c)
			fs := service.NewMockFileService(c)
			te := NewMocktemplateExecutor(c)
			handler := Handler{
				templates: te,
				service:   service.Service{Files: fs, Session: ss},
			}

			tc.sessionServiceBehaviour(ss, tc.sessionID)
			tc.fileServiceBehaviour(fs, &service.Session{})
			tc.templateBehavior(te)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/cut", nil)
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			ctxr := tc.requestModification(r, tc.sessionID)
			handler.MainPage(w, ctxr)
		})
	}
}

func TestRouter_DownloadFile(t *testing.T) {
	testCases := []struct {
		name                    string
		sessionID               string
		requestModification     func(r *http.Request, sessionID string) *http.Request
		sessionServiceBehaviour func(mss *service.MockSessionService, sessionID string)
		fileServiceBehaviour    func(mfs *service.MockFileService, session *service.Session)
		templateBehavior        func(te *MocktemplateExecutor)
		responseCode            int
	}{
		{
			name: "ok",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestRouter_UploadFile(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestRouter_DeleteFile(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
