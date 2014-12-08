package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const TWITTER_API_URL = "https://api.twitter.com/1.1"

type TwitterApi struct {
	BaseUrl     string
	AccessToken string
}

func NewTwitterApi(baseUrl, accesToken string) *TwitterApi {
	return &TwitterApi{baseUrl, accesToken}
}

func (t *TwitterApi) encodeParams(params map[string]string) string {

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return q.Encode()
}

func (t *TwitterApi) asCommaSeparatedString(lst []uint64) string {
	lstAsString := make([]string, len(lst))
	for i, n := range lst {
		lstAsString[i] = fmt.Sprintf("%v", n)
	}
	return strings.Join(lstAsString, ",")
}

func (t *TwitterApi) createGetPathAndParams(apiEndpoint string, params map[string]string) string {
	u, err := url.Parse(apiEndpoint)
	if err != nil {
		panic(err)
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (t *TwitterApi) GetBase64EncodedBearerTokenCredentials(consumerKey, consumerSecret string) string {
	data := []byte(consumerKey + ":" + consumerSecret)
	return base64.URLEncoding.EncodeToString(data)
}

func (t *TwitterApi) Get(path_ string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", t.BaseUrl+path_, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	req.Header.Set("content-type", "application/json; charset=utf-8")
	return http.DefaultClient.Do(req)
}

func (t *TwitterApi) Post(path_ string, body string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", t.BaseUrl+path_, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	return http.DefaultClient.Do(req)
}

func (t *TwitterApi) GetAndDeserialize(path string, params map[string]string, v interface{}) (err error) {
	defer func() {
		if err == io.EOF {
			err = nil
		}
	}()
	if r, err := t.Get(t.createGetPathAndParams(path, params)); err != nil {
		return err
	} else if r.StatusCode/100 != 2 {
		if errMsg, err := ioutil.ReadAll(r.Body); err != nil {
			return errors.New("error getting endpoint and cannot deserialize error message")
		} else {
			return NewTwitterErr(string(errMsg), r.StatusCode)
		}
	} else {
		d := json.NewDecoder(r.Body)
		return d.Decode(v)
	}
}

func (t *TwitterApi) PostAndDeserialize(path string, params map[string]string, v interface{}) (err error) {
	defer func() {
		if err == io.EOF {
			err = nil
		}
	}()
	if r, err := t.Post(path, t.encodeParams(params)); err != nil {
		return err
	} else if r.StatusCode/100 != 2 {
		if errMsg, err := ioutil.ReadAll(r.Body); err != nil {
			return errors.New("error getting endpoint and cannot deserialize error message")
		} else {
			return NewTwitterErr(string(errMsg), r.StatusCode)
		}
	} else {
		d := json.NewDecoder(r.Body)
		return d.Decode(v)
	}
}

func (t *TwitterApi) GetTwitterIdByScreenName(sceenName string) (id uint64, err error) {
	ids := make([]*idHolder, 0, 1)
	params := map[string]string{"screen_name": sceenName, "include_entities": "id"}
	apiPath := "/users/lookup.json"
	if err := t.GetAndDeserialize(apiPath, params, &ids); err != nil {
		return 0, err
	} else if len(ids) < 1 {
		return 0, errors.New("cannot find user with screen name " + sceenName)
	}
	return ids[0].Id, nil
}

func (t *TwitterApi) GetFollowerByCursor(screenName, cursor string) <-chan *FollowerList {
	followerListC := make(chan *FollowerList)
	go func() {
		params := map[string]string{"screen_name": screenName, "count": "200", "skip_status": "true", "cursor": cursor}
		apiPath := "/followers/list.json"
		followers := new(FollowerList)
		if err := t.GetAndDeserialize(apiPath, params, followers); err == nil {
			followerListC <- followers
		} else if twitterErr, ok := err.(*TwitterErr); ok && twitterErr.Status == 429 {
			log.Println("api limit reached, need to sleep")
			time.Sleep(5 * 60 * time.Second)
			followerListC <- <-t.GetFollowerByCursor(screenName, cursor)
		} else {
			log.Println(err)
			followerListC <- nil
		}
	}()
	return followerListC
}

func (t *TwitterApi) GetFollowerIdsByCursor(screenName, cursor string) <-chan *FollowerIDList {
	followerListC := make(chan *FollowerIDList)

	go func() {
		params := map[string]string{"screen_name": screenName, "count": "5000", "cursor": cursor}
		apiPath := "/followers/ids.json"
		followers := new(FollowerIDList)
		if err := t.GetAndDeserialize(apiPath, params, followers); err == nil {
			followerListC <- followers
		} else if twitterErr, ok := err.(*TwitterErr); ok && twitterErr.Status == 429 {
			log.Println("api limit reached, need to sleep")
			time.Sleep(5 * 60 * time.Second)
			followerListC <- <-t.GetFollowerIdsByCursor(screenName, cursor)
		} else {
			log.Println(err)
			followerListC <- nil
		}
	}()

	return followerListC
}

func (t *TwitterApi) GetScreenNameOfUsersByIds(ids []uint64) <-chan string {
	if len(ids) >= 100 {
		log.Println("GetScreenNameOfUsersByIds received a list of more that 100 ids.  This is not supported by twitter")
	}
	screenNameC := make(chan string)
	go func() {
		path := "/users/lookup.json"
		params := map[string]string{"user_id": t.asCommaSeparatedString(ids)}
		users := make([]*User, 0, len(ids))
		if err := t.PostAndDeserialize(path, params, &users); err == nil {
			for _, user := range users {
				screenNameC <- user.ScreenName
			}
		} else if twitterErr, ok := err.(*TwitterErr); ok && twitterErr.Status == 429 {
			log.Println("api limit reached, need to sleep")
			time.Sleep(5 * 60 * time.Second)
			screenNameC <- <-t.GetScreenNameOfUsersByIds(ids)
		} else {
			log.Println(err)
			screenNameC <- ""
		}
		close(screenNameC)

	}()
	return screenNameC
}
