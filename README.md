# AWS外部鉴权转接器

## 支持
| 云商 | 方法 |
| :- | :- |
| 腾讯云 | [实例CAM角色](https://cloud.tencent.com/document/product/213/47668) |
| 阿里云 | [实例RAM角色](https://help.aliyun.com/zh/ecs/user-guide/attach-an-instance-ram-role-to-an-ecs-instance) |

## 用法
```sh
usage: external-cred-adapter <command> [<args> ...]

Commands:
tencentcloud role [<roleName>] [{security-credentials,service-role-security-credentials}]
    请求腾讯云角色临时密钥

aliyun role [<roleName>]
    请求阿里云角色临时密钥
```

## 配置方法
```ini
# ~/.aws/config
[default]
credential_process = /path/to/external-cred-adapter tencentcloud role

[profile aliyun]
credential_process = /path/to/external-cred-adapter aliyun role
```