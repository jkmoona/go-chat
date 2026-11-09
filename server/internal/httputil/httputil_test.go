package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q, want application/json", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["key"] != "value" {
		t.Fatalf("body key = %q, want %q", got["key"], "value")
	}
}

func TestErrorJSON(t *testing.T) {
	w := httptest.NewRecorder()
	ErrorJSON(w, http.StatusBadRequest, "bad input")

	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["error"] != "bad input" {
		t.Fatalf("error = %q, want %q", got["error"], "bad input")
	}
}

func TestMsgJSON(t *testing.T) {
	w := httptest.NewRecorder()
	MsgJSON(w, http.StatusOK, "message", "done")

	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["message"] != "done" {
		t.Fatalf("message = %q, want %q", got["message"], "done")
	}
}

func TestDecodeAndValidate(t *testing.T) {
	type req struct {
		Name string `json:"name" validate:"required,min=2"`
	}

	t.Run("valid", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"ab"}`))
		var dst req
		if err := DecodeAndValidate(r, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.Name != "ab" {
			t.Fatalf("name = %q, want %q", dst.Name, "ab")
		}
	})

	t.Run("missing required", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		var dst req
		err := DecodeAndValidate(r, &dst)
		if err == nil {
			t.Fatal("expected validation error")
		}
		var ve validator.ValidationErrors
		if !errors.As(err, &ve) {
			t.Fatalf("expected ValidationErrors, got %T", err)
		}
	})

	t.Run("too short", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"a"}`))
		var dst req
		err := DecodeAndValidate(r, &dst)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("bad json", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad`))
		var dst req
		if err := DecodeAndValidate(r, &dst); err == nil {
			t.Fatal("expected decode error")
		}
	})
}

func TestSetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	SetCookie(w, "token", "abc", 900, true, true)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	c := cookies[0]
	if c.Name != "token" || c.Value != "abc" {
		t.Fatalf("cookie = %s=%s, want token=abc", c.Name, c.Value)
	}
	if c.MaxAge != 900 {
		t.Fatalf("max-age = %d, want 900", c.MaxAge)
	}
	if !c.Secure || !c.HttpOnly {
		t.Fatal("expected Secure and HttpOnly")
	}
	if c.Path != "/" {
		t.Fatalf("path = %q, want /", c.Path)
	}
}

func TestReadCookie(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "token", Value: "abc"})

	val, err := ReadCookie(r, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "abc" {
		t.Fatalf("value = %q, want %q", val, "abc")
	}

	_, err = ReadCookie(r, "missing")
	if err == nil {
		t.Fatal("expected error for missing cookie")
	}
}

func TestBindingError(t *testing.T) {
	type req struct {
		Name string `validate:"required"`
		Age  int    `validate:"min=1"`
	}

	err := validate.Struct(&req{})
	msg := BindingError(err)
	if msg != "name is required" {
		t.Fatalf("msg = %q, want %q", msg, "name is required")
	}

	msg = BindingError(errors.New("some other error"))
	if msg != "invalid request" {
		t.Fatalf("msg = %q, want %q", msg, "invalid request")
	}
}
