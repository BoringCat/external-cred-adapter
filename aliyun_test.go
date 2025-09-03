package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
)

var (
	aliyunToken string
)

func handleToken(w http.ResponseWriter, r *http.Request) {
	data := make([]byte, 64)
	rand.Read(data)
	aliyunToken = base64.RawStdEncoding.EncodeToString(data)
	fmt.Fprint(w, aliyunToken)
}
func handleRole(cred *aliyunRoleCredential) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == aliyunRamPath {
			fmt.Fprint(w, "test-credential-role")
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(cred); err != nil {
			fmt.Println(err)
		}
	}
}

func TestAliyunCredSuccess(t *testing.T) {
	cred := &aliyunRoleCredential{}
	if err := faker.FakeData(cred); err != nil {
		t.Fatal(err)
	}
	cred.Code = "Success"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, aliyunTokenPath) {
			handleToken(w, r)
		} else if strings.HasPrefix(r.URL.Path, aliyunRamPath) {
			handleRole(cred)(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	aliyunMetadataUrl = ts.URL
	fn := AliyunRoleCredential
	ctx, cancel := context.WithTimeout(context.TODO(), timedOut)
	defer cancel()
	resp, err := fn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessKeyId != cred.AccessKeyId {
		t.Fatal("AccessKeyId not match AccessKeyId")
	}
	if resp.SecretAccessKey != cred.AccessKeySecret {
		t.Fatal("SecretAccessKey not match AccessKeySecret")
	}
	if resp.SessionToken != cred.SecurityToken {
		t.Fatal("SessionToken not match SecurityToken")
	}
	if !resp.Expiration.Equal(cred.Expiration) {
		t.Fatal("Expiration not match Expiration")
	}
}

func TestAliyunCredFailed(t *testing.T) {
	cred := &aliyunRoleCredential{}
	if err := faker.FakeData(cred); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, aliyunTokenPath) {
			handleToken(w, r)
		} else if strings.HasPrefix(r.URL.Path, aliyunRamPath) {
			handleRole(cred)(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	aliyunMetadataUrl = ts.URL
	fn := AliyunRoleCredential
	ctx, cancel := context.WithTimeout(context.TODO(), timedOut)
	defer cancel()
	_, err := fn(ctx)
	if err == nil {
		t.Fatal("err is nil")
	}
}
