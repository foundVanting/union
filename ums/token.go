package ums

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/tiantour/fetch"
	"github.com/tiantour/imago"
	"github.com/tiantour/rsae"
	"github.com/tiantour/union/x/cache"
)

// Token token
type Token struct{}

// NewToken new token
func NewToken() *Token {
	return &Token{}
}

// Access access token
func (t *Token) Access() (string, error) {
	token, ok := cache.NewString().Get(AppID)
	if ok && token != "" {
		return token.(string), nil
	}

	data := &Request{
		AppID:      AppID,
		Timestamp:  time.Now().Format("20060102150405"),
		Nonce:      imago.NewRandom().Text(32),
		SignMethod: "SHA256",
	}
	sign := rsae.NewSHA().SHA256(AppID + data.Timestamp + data.Nonce + AppKey)
	data.Signature = string(hex.EncodeToString(sign))

	result, err := t.do(data)
	if err != nil {
		return "", err
	}

	_ = cache.NewString().Set(AppID, result.AccessToken, 1, 7200*time.Second)
	return result.AccessToken, nil
}

// do do
func (t *Token) do(args *Request) (*Response, error) {
	body, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Add("Accept", "application/json")
	header.Add("Content-Type", "application/json;charset=utf-8")
	body, err = fetch.Cmd(&fetch.Request{
		Method: "POST",
		URL:    "https://api-mop.chinaums.com/v1/token/access",
		Body:   body,
		Header: header,
	})
	if err != nil {
		return nil, err
	}

	result := Response{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != "0000" {
		return nil, errors.New(result.ErrInfo)
	}
	return &result, err
}
