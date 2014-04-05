// Package github implements a simple client to consume gitlab API.
package gogitlab

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	dasboard_feed_path = "/dashboard.atom"
)

type Gitlab struct {
	BaseUrl      string
	ApiPath      string
	RepoFeedPath string
	Token        string
	Client       *http.Client
}

const (
	dateLayout  = "2006-01-02T15:04:05-07:00"
	session_url = "/session" // Login to get private token

)

type LoginInfo struct {
	Id           int    `json:"id,omitempty"`
	Username     string `json:"username,omitempty"`
	Email        string `json:"email,omitempty"`
	Name         string `json:"name,omitempty"`
	State        string `json:"state,omitempty"`
	CreatedAtRow string `json:"created_at,omitempty"`
	CreatedAt    time.Time
	externUid    string `json:"extern_uid,omitempty"`
	Provider     string `json:"provider,omitempty"`
	Token        string `json:"private_token,omitempty"`
	IsAdmin      bool   `json:"is_admin,omitempty"`
}

func NewGitlab(baseUrl, apiPath, token string) *Gitlab {

	client := &http.Client{}

	return &Gitlab{
		BaseUrl: baseUrl,
		ApiPath: apiPath,
		Token:   token,
		Client:  client,
	}
}

func NewGitlabByLogin(baseUrl, apiPath, username, password string) (*Gitlab, error) {

	client := &http.Client{}
	gitlab := &Gitlab{
		BaseUrl: baseUrl,
		ApiPath: apiPath,
		Client:  client,
	}

	loginUrl := gitlab.BaseUrl + gitlab.ApiPath + session_url

	data := url.Values{}
	data.Set("login", username)
	data.Add("password", password)

	contents, err := gitlab.buildAndExecRequest("POST", loginUrl, []byte(data.Encode()))
	var loginInfo *LoginInfo
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &loginInfo)
	if err != nil {
		return nil, err
	}
	gitlab.Token = loginInfo.Token
	return gitlab, nil
}

func (g *Gitlab) ResourceUrl(url string, params map[string]string) string {

	if params != nil {
		for key, val := range params {
			url = strings.Replace(url, key, val, -1)
		}
	}

	url = g.BaseUrl + g.ApiPath + url + "?private_token=" + g.Token

	return url
}

func (g *Gitlab) buildAndExecRequest(method, url string, body []byte) ([]byte, error) {

	var req *http.Request
	var err error

	if body != nil {
		reader := bytes.NewReader(body)
		req, err = http.NewRequest(method, url, reader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		panic("Error while building gitlab request")
	}

	resp, err := g.Client.Do(req)
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s", err)
	}

	if resp.StatusCode >= 400 {
		err = errors.New("*Gitlab.buildAndExecRequest failed: " + resp.Status)
	}

	return contents, err
}
