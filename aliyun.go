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
	aliyunMetadataUrl = "http://100.100.100.200"
	aliyunRamPath     = "/latest/meta-data/ram"
	aliyunTokenPath   = "/latest/api/token"
	aliyunRoleName    string
)

func getAliyunToken() (string, error) {
	req, err := http.NewRequest(http.MethodPut, aliyunMetadataUrl+aliyunTokenPath, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-aliyun-ecs-metadata-token-ttl-seconds", "10")
	result, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getAliyunRoleName(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, aliyunMetadataUrl+aliyunRamPath, nil)
	req.Header.Add("X-aliyun-ecs-metadata-token", token)
	result, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}
	if result.StatusCode != 200 {
		return "", fmt.Errorf("获取角色名称返回异常: %s\n\turl: %s\n%s", result.Status, result.Request.URL.String(), string(data))
	}
	for line := range bytes.Lines(data) {
		return string(line), nil
	}
	return "", io.EOF
}

type aliyunRoleCredential struct {
	AccessKeyId     string
	AccessKeySecret string
	Expiration      time.Time
	SecurityToken   string
	LastUpdated     time.Time
	Code            string
}

func AliyunRoleCredential() (*externalProcessCredentialResult, error) {
	token, err := getAliyunToken()
	if err != nil {
		return nil, err
	}
	if aliyunRoleName == "" {
		var err error
		aliyunRoleName, err = getAliyunRoleName(token)
		if err != nil {
			panic(err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, aliyunMetadataUrl+aliyunRamPath+"/security-credentials/"+aliyunRoleName, nil)
	req.Header.Add("X-aliyun-ecs-metadata-token", token)
	result, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	if result.StatusCode != 200 {
		body, _ := io.ReadAll(result.Body)
		return nil, fmt.Errorf("获取临时密钥返回异常: %s\n\turl: %s\n%s", result.Status, result.Request.URL.String(), string(body))
	}
	cred := new(aliyunRoleCredential)
	if err = json.NewDecoder(result.Body).Decode(cred); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	if cred.Code != "Success" {
		return nil, fmt.Errorf("get credential failed: %s", cred.Code)
	}
	return &externalProcessCredentialResult{
		Version:         1,
		AccessKeyId:     cred.AccessKeyId,
		SecretAccessKey: cred.AccessKeySecret,
		SessionToken:    cred.SecurityToken,
		Expiration:      cred.Expiration,
	}, nil
}
