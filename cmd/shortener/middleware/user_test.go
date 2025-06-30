package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithUserIDMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		initialCookie   *http.Cookie
		expectNewCookie bool
	}{
		{
			name:            "No initial cookie",
			initialCookie:   nil,
			expectNewCookie: true,
		},
		{
			name:            "Empty initial cookie",
			initialCookie:   &http.Cookie{Name: userIDCookieName, Value: ""},
			expectNewCookie: true,
		},
		{
			name:            "Existing cookie present",
			initialCookie:   &http.Cookie{Name: userIDCookieName, Value: "existing-id"},
			expectNewCookie: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := WithUserID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := FromContext(r.Context())
				if !ok || userID == "" {
					t.Errorf("Expected userID to be set in context, got %v", userID)
				}
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.initialCookie != nil {
				req.AddCookie(tc.initialCookie)
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}(res.Body)
			cookies := res.Cookies()

			if tc.expectNewCookie {
				if len(cookies) == 0 {
					t.Errorf("Expected a new cookie to be set")
				} else if cookies[0].Name != userIDCookieName || cookies[0].Value == "" {
					t.Errorf("Unexpected cookie value: got %v", cookies[0].Value)
				}
			} else {
				if len(cookies) > 0 {
					t.Errorf("Expected no new cookie, but got one")
				}
			}
		})
	}
}
