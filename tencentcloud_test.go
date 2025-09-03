package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faker/faker/v4"
)

func TestTencentCloudCredSuccess(t *testing.T) {
	cred := &tencentCloudRoleCredential{}
	if err := faker.FakeData(cred); err != nil {
		t.Fatal(err)
	}
	cred.Code = "Success"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fmt.Fprint(w, "test-credential-role")
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(cred); err != nil {
			fmt.Println(err)
		}
	}))
	defer ts.Close()
	tencentCloudCamUrl = ts.URL
	fn := TencentCloudRoleCredential
	ctx, cancel := context.WithTimeout(context.TODO(), timedOut)
	defer cancel()
	resp, err := fn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessKeyId != cred.TmpSecretId {
		t.Fatal("AccessKeyId not match TmpSecretId")
	}
	if resp.SecretAccessKey != cred.TmpSecretKey {
		t.Fatal("SecretAccessKey not match TmpSecretKey")
	}
	if resp.SessionToken != cred.Token {
		t.Fatal("SessionToken not match Token")
	}
	if !resp.Expiration.Equal(cred.Expiration) {
		t.Fatal("Expiration not match Expiration")
	}
}

func TestTencentCloudCredFailed(t *testing.T) {
	cred := &tencentCloudRoleCredential{}
	if err := faker.FakeData(cred); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fmt.Fprint(w, "test-credential-role")
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(cred); err != nil {
			fmt.Println(err)
		}
	}))
	defer ts.Close()
	tencentCloudCamUrl = ts.URL
	fn := TencentCloudRoleCredential
	ctx, cancel := context.WithTimeout(context.TODO(), timedOut)
	defer cancel()
	_, err := fn(ctx)
	if err == nil {
		t.Fatal("err is nil")
	}
}
