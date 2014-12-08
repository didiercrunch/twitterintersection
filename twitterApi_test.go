package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestcreateGetPathAndParams(t *testing.T) {
	tw := NewTwitterApi("twitter.com", "access_token")
	p := map[string]string{"hello": "world"}
	if u := tw.createGetPathAndParams("bob/foo", p); u != "bob/foo?hello=world" {
		t.Error("bad url", u)
	}
}

func TestGetBase64EncodedBearerTokenCredentials(t *testing.T) {
	key := "xvz1evFS4wEEPTGEFPHBog"
	secret := "L8qq9PZyRg6ieKGEKhZolGC0vJWLw8iEJ88DRdyOg"
	expt := "eHZ6MWV2RlM0d0VFUFRHRUZQSEJvZzpMOHFxOVBaeVJnNmllS0dFS2hab2xHQzB2SldMdzhpRUo4OERSZHlPZw=="
	tapi := new(TwitterApi)
	if res := tapi.GetBase64EncodedBearerTokenCredentials(key, secret); res != expt {
		t.Error("expected: ", expt, "\nbut received:", res)
	}
}

func TestAsCommaSeparatedString(t *testing.T) {
	tw := new(TwitterApi)
	if s := tw.asCommaSeparatedString([]uint64{10765432100123456789, 78}); s != "10765432100123456789,78" {
		t.Error(s)
	}
}

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("bad request verb")
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer access_token" {
			t.Error("bad authorization in header :", auth)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
			t.Error("bad content-type in header ", contentType)
		}
	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")

	if res, err := tw.Get("/"); err != nil {
		t.Error(err)
	} else {
		ioutil.ReadAll(res.Body)
	}
}

func TestPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("bad request verb")
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer access_token" {
			t.Error("bad authorization in header :", auth)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
			t.Error("bad content-type in header ", contentType)
		}
		if body, err := ioutil.ReadAll(r.Body); err != nil {
			t.Error(err)
		} else if string(body) != `{"data": "is_hugue"}` {
			t.Error("bad body", string(body))
		}
	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")

	if res, err := tw.Post("/", `{"data": "is_hugue"}`); err != nil {
		t.Error(err)
	} else {
		ioutil.ReadAll(res.Body)
	}
}

func TestGetAndDeserializeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error": "bar"}`)
	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")
	m := make(map[string]string)
	if err := tw.GetAndDeserialize("/foo", map[string]string{}, &m); err == nil {
		t.Error("shoud have an error message here")
	} else if err.Error() != `{"error": "bar"}` {
		t.Error("error message should propagate here")
	} else if terr, ok := err.(*TwitterErr); !ok {
		t.Error("the error should be a twitter error")
	} else if terr.Status != 404 {
		t.Error("bad status")
	}
}

func TestGetAndDeserialize(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("foo") != "bar" {
			t.Error("bad params")
		}
		fmt.Fprint(w, `{"foo": "bar"}`)
	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")
	m := make(map[string]string)
	if err := tw.GetAndDeserialize("/foo", map[string]string{"foo": "bar"}, &m); err != nil {
		t.Error(err)
	}
	if m["foo"] != "bar" {
		t.Fail()
	}
}

func TestPostAndDeserialize(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, err := ioutil.ReadAll(r.Body); err != nil {
			t.Error(err)
		} else if string(body) != `val=7` {
			t.Error("bad body")
		}
		fmt.Fprint(w, `{"foo": "bar"}`)
	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")
	m := make(map[string]string)

	if err := tw.PostAndDeserialize("/foo", map[string]string{"val": "7"}, &m); err != nil {
		t.Error(err)
	}
	if m["foo"] != "bar" {
		t.Fail()
	}
}

func TestGetTwitterIdByScreenName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if url_ := r.URL.String(); url_ != "/users/lookup.json?include_entities=id&screen_name=bobLeChef" {
			t.Error("bad url :", url_)
		}
		fmt.Fprint(w, `[{"id": 1492}]`)

	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")

	if id, err := tw.GetTwitterIdByScreenName("bobLeChef"); err != nil {
		t.Error(err)
	} else if id != 1492 {
		t.Error("found bad id")
	}
}

func TestGetFollowerByCursor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if url_ := r.URL.String(); url_ != "/followers/list.json?count=200&cursor=89&screen_name=bobLeChef&skip_status=true" {
			t.Error("bad url :", url_)
		}
		fmt.Fprint(w, `{"users": [{"id": 1492, "screen_name": "bob_le_chef"}], "next_cursor_str": "1793"}`)
	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")

	followersListC := tw.GetFollowerByCursor("bobLeChef", "89")
	followers := <-followersListC

	if followers.NextCursor != "1793" {
		t.Error("bad cursor")
	} else if len(followers.Followers) != 1 {
		t.Error("bad number of followers")
	} else if user := followers.Followers[0]; !reflect.DeepEqual(user, &User{"bob_le_chef", 1492}) {
		t.Error("bad user", user)
	}
}

func TestGetFollowerIdsByCursor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if url_ := r.URL.String(); url_ != "/followers/ids.json?count=5000&cursor=89&screen_name=bobLeChef" {
			t.Error("bad url :", url_)
		}
		fmt.Fprint(w, `{"ids": [1492], "next_cursor_str": "1793"}`)
	}))
	defer ts.Close()
	tw := NewTwitterApi(ts.URL, "access_token")

	followersListC := tw.GetFollowerIdsByCursor("bobLeChef", "89")

	followers := <-followersListC

	if followers.NextCursor != "1793" {
		t.Error("bad cursor")
	} else if len(followers.Followers) != 1 {
		t.Error("bad number of followers", len(followers.Followers))
	} else if id := followers.Followers[0]; id != 1492 {
		t.Error("bad id")
	}
}
