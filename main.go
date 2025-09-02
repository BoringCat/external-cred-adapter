package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
)

var (
	version      string
	buildDate    string
	commit       string
	printVersion bool
)

type getCredentialFn func() (*externalProcessCredentialResult, error)

func parseArgs() string {
	app := kingpin.New("external-cred-adapter", "转换各种云的临时密钥到AWS外部验证")
	app.Command("version", "打印版本号")
	qcloud := app.Command("tencentcloud", "腾讯云").Alias("qcloud")
	qcloud_role := qcloud.Command("role", "请求腾讯云角色临时密钥")
	qcloud_role.Arg("roleName", "角色").StringVar(&tencentCloudRoleName)
	qcloud_role.Arg("type", "角色类型").Default("security-credentials").EnumVar(&tencentCloudRoleType, "security-credentials", "service-role-security-credentials")

	aliyun := app.Command("aliyun", "阿里云").Alias("alibaba")
	aliyun.Command("role", "请求阿里云角色临时密钥").Arg("roleName", "角色").StringVar(&tencentCloudRoleName)
	return kingpin.MustParse(app.Parse(os.Args[1:]))
}

func main() {
	var fn getCredentialFn
	switch parseArgs() {
	case "tencentcloud role":
		fn = TencentCloudRoleCredential
	case "aliyun role":
		fn = AliyunRoleCredential
	case "version":
		fmt.Printf("external-cred-adapter, version %s (date: %s, revision: %s)\n", version, buildDate, commit)
		fmt.Printf("  go version:\t%s\n", runtime.Version())
		fmt.Printf("  platform:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}
	if fn == nil {
		return
	}
	cred, err := fn()
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(cred)
}
