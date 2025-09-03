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
	tencentCloudCamUrl   = "http://metadata.tencentyun.com/latest/meta-data/cam"
	tencentCloudRoleName string
	tencentCloudRoleType string
)

func getTencentCloudRoleName(ctx context.Context) (string, error) {
	req, err := http.NewRequest(http.MethodGet, tencentCloudCamUrl+"/"+tencentCloudRoleType, nil)
	if err != nil {
		return "", err
	}
	debugPrintln("[腾讯云]", "[获取角色名称]", "请求URL:", req.URL.String())
	result, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	debugPrintln("[腾讯云]", "[获取角色名称]", "响应状态码:", result.Status)
	debugPrintln("[腾讯云]", "[获取角色名称]", "响应头:", result.Header)
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}
	debugPrintln("[腾讯云]", "[获取角色名称]", "响应体:", string(data))
	if result.StatusCode != 200 {
		return "", fmt.Errorf("获取角色名称返回异常: %s", result.Status)
	}
	for line := range bytes.Lines(data) {
		debugPrintln("[腾讯云]", "[获取角色名称]", "返回值", string(line))
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

func TencentCloudRoleCredential(ctx context.Context) (*externalProcessCredentialResult, error) {
	if tencentCloudRoleName == "" {
		var err error
		tencentCloudRoleName, err = getTencentCloudRoleName(ctx)
		if err != nil {
			panic(err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, tencentCloudCamUrl+"/"+tencentCloudRoleType+"/"+tencentCloudRoleName, nil)
	if err != nil {
		return nil, err
	}
	debugPrintln("[腾讯云]", "[获取临时密钥]", "请求URL:", req.URL.String())
	result, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	debugPrintln("[腾讯云]", "[获取临时密钥]", "响应状态码:", result.Status)
	debugPrintln("[腾讯云]", "[获取临时密钥]", "响应头:", result.Header)
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	debugPrintln("[腾讯云]", "[获取临时密钥]", "响应体:", string(data))
	if result.StatusCode != 200 {
		return nil, fmt.Errorf("获取临时密钥返回异常: %s", result.Status)
	}
	cred := new(tencentCloudRoleCredential)
	if err = json.Unmarshal(data, cred); err != nil {
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
