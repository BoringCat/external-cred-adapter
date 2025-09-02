package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	tencentCloudCamUrl   = "http://metadata.tencentyun.com/latest/meta-data/cam"
	tencentCloudRoleName string
	tencentCloudRoleType string
)

func getTencentCloudRoleName() (string, error) {
	result, err := http.Get(tencentCloudCamUrl + "/" + tencentCloudRoleType)
	if err != nil {
		return "", err
	}
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}
	for line := range bytes.Lines(data) {
		return string(line), nil
	}
	return "", io.EOF
}

type tencentCloudRoleCredential struct {
	TmpSecretId  string
	TmpSecretKey string
	ExpiredTime  int
	Expiration   time.Time
	Token        string
	Code         string
}

func TencentCloudRoleCredential() (*externalProcessCredentialResult, error) {
	if tencentCloudRoleName == "" {
		var err error
		tencentCloudRoleName, err = getTencentCloudRoleName()
		if err != nil {
			panic(err)
		}
	}
	url := tencentCloudCamUrl + "/" + tencentCloudRoleType + "/" + tencentCloudRoleName
	result, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	cred := new(tencentCloudRoleCredential)
	if err = json.NewDecoder(result.Body).Decode(cred); err != nil {
		return nil, err
	}
	if cred.Code != "Success" {
		return nil, fmt.Errorf("get credential failed: %s", cred.Code)
	}
	return &externalProcessCredentialResult{
		Version:         1,
		AccessKeyId:     cred.TmpSecretId,
		SecretAccessKey: cred.TmpSecretKey,
		SessionToken:    cred.Token,
		Expiration:      cred.Expiration,
	}, nil
}
