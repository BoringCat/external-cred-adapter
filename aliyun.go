package main

import (
	"bytes"
	"context"
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

func getAliyunToken(ctx context.Context) (string, error) {
	req, err := http.NewRequest(http.MethodPut, aliyunMetadataUrl+aliyunTokenPath, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-aliyun-ecs-metadata-token-ttl-seconds", "10")
	debugPrintln("[阿里云]", "[获取访问Token]", "请求URL:", req.URL.String())
	debugPrintln("[阿里云]", "[获取访问Token]", "请求头:", req.Header)
	result, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	debugPrintln("[阿里云]", "[获取访问Token]", "响应状态码:", result.Status)
	debugPrintln("[阿里云]", "[获取访问Token]", "响应头:", result.Header)
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	debugPrintln("[阿里云]", "[获取访问Token]", "响应体:", string(data))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getAliyunRoleName(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, aliyunMetadataUrl+aliyunRamPath, nil)
	req.Header.Add("X-aliyun-ecs-metadata-token", token)
	debugPrintln("[阿里云]", "[获取角色名称]", "请求URL:", req.URL.String())
	debugPrintln("[阿里云]", "[获取角色名称]", "请求头:", req.Header)
	result, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	debugPrintln("[阿里云]", "[获取角色名称]", "响应状态码:", result.Status)
	debugPrintln("[阿里云]", "[获取角色名称]", "响应头:", result.Header)
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	debugPrintln("[阿里云]", "[获取角色名称]", "响应体:", string(data))
	if err != nil {
		return "", err
	}
	if result.StatusCode != 200 {
		return "", fmt.Errorf("获取角色名称返回异常: %s", result.Status)
	}
	for line := range bytes.Lines(data) {
		debugPrintln("[阿里云]", "[获取角色名称]", "返回值", string(line))
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

func AliyunRoleCredential(ctx context.Context) (*externalProcessCredentialResult, error) {
	token, err := getAliyunToken(ctx)
	if err != nil {
		return nil, err
	}
	if aliyunRoleName == "" {
		var err error
		aliyunRoleName, err = getAliyunRoleName(ctx, token)
		if err != nil {
			panic(err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, aliyunMetadataUrl+aliyunRamPath+"/security-credentials/"+aliyunRoleName, nil)
	req.Header.Add("X-aliyun-ecs-metadata-token", token)
	debugPrintln("[阿里云]", "[获取临时密钥]", "请求URL:", req.URL.String())
	debugPrintln("[阿里云]", "[获取临时密钥]", "请求头:", req.Header)
	result, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	debugPrintln("[阿里云]", "[获取临时密钥]", "响应状态码:", result.Status)
	debugPrintln("[阿里云]", "[获取临时密钥]", "响应头:", result.Header)
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	debugPrintln("[阿里云]", "[获取临时密钥]", "响应体:", string(data))
	if result.StatusCode != 200 {
		return nil, fmt.Errorf("获取临时密钥返回异常: %s", result.Status)
	}
	cred := new(aliyunRoleCredential)
	if err = json.Unmarshal(data, cred); err != nil {
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
