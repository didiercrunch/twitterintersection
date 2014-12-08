package main

import "testing"

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
