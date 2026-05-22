# aitk SSH Examples

可直接运行的 `aitk SSH_` 命令示例。支持三种输入模式: hex、明文JSON、stdin。

## 使用方法

1. 将命令中的 `YOUR_HOST` / `YOUR_PASSWORD` 替换为实际值
2. 直接粘贴到终端运行

三种输入模式示例:

```bash
# Hex模式 (AI agent推荐):
aitk SSH_7b22686f7374223a...

# 明文JSON模式 (人类友好):
aitk 'SSH_{"host":"1.2.3.4","port":22,"user":"root","password":"secret","action":"cmd","cmd":"hostname"}'

# Stdin模式 (管道组合):
echo '{"host":"1.2.3.4","port":22,"user":"root","password":"secret","action":"cmd","cmd":"hostname"}' | aitk SSH
```

## 示例列表

| 文件 | 说明 |
|------|------|
| `01_cmd.txt` | 执行远程命令 (hex/明文/stdin三种模式) |
| `02_cmd_timeout.txt` | 命令超时控制 |
| `03_cmd_file.txt` | 从文件执行多条命令 |
| `04_upload.txt` | 上传文件 |
| `05_download.txt` | 下载文件 |
| `06_upload_atomic.txt` | 原子上传（临时文件+重命名） |
| `07_mkdir.txt` | 创建远程目录 |
| `08_remove.txt` | 删除远程文件/目录 |
| `09_chmod.txt` | 修改文件权限 |
| `10_move.txt` | 移动/重命名远程文件 |
| `11_deploy.txt` | 多步骤部署计划 |
| `12_sync_push.txt` | 同步推送 (local → remote) |
| `13_sync_pull.txt` | 同步拉取 (remote → local) |
| `14_sync_bidirectional.txt` | 双向同步 + 冲突策略 |
| `15_sync_single_file.txt` | 单文件同步 |
| `16_key_auth.txt` | 私钥认证 |
| `17_errors.txt` | 错误场景示例 |
| `18_deploy_advanced.txt` | 高级部署 (多步骤类型、容忍错误、连接超时) |
| `deploy_plan_example.json` | 部署计划 JSON 示例文件 |

## SSH Actions 参考

| Action | 必填字段 | 说明 |
|--------|---------|------|
| `cmd` | `cmd` 或 `cmd_file` | 执行远程命令 |
| `upload` | `local_path`, `remote_path` | 上传文件 |
| `download` | `local_path`, `remote_path` | 下载文件 |
| `upload_atomic` | `local_path`, `remote_path` | 原子上传 |
| `mkdir` | `remote_path` | 创建目录 |
| `remove` | `remote_path` | 删除文件/目录 |
| `chmod` | `remote_path`, `mode` | 修改权限 |
| `move` | `source`, `target` | 移动/重命名 |
| `deploy` | `plan` 或 `plan_json` | 多步骤部署 |
| `sync` | `local_path`, `remote_path`, `direction` | 文件同步 |

## 认证方式

| 方式 | 字段 |
|------|------|
| 密码 | `password` |
| 密钥 | `key` (路径) |
| 密钥+密码 | `key` + `key_passphrase` |

安全选项: `strict_host_key` (bool), `known_hosts` (路径), `timeout` (连接超时, 如 "10s"), `cmd_timeout` (命令超时, 如 "30s")

## Sync 方向

| 方向 | 说明 |
|------|------|
| `push` | 本地 → 远程 |
| `pull` | 远程 → 本地 |
| `bidirectional` | 双向同步 |

## Sync 冲突策略

| 策略 | 行为 |
|------|------|
| `fail_on_conflict` | 报告冲突，跳过文件（默认） |
| `newer_wins` | 较新的文件覆盖较旧的 |
| `local_wins` | 本地覆盖远程 |
| `remote_wins` | 远程覆盖本地 |

## Deploy 步骤类型

| 类型 | 必填字段 | 说明 |
|------|---------|------|
| `cmd` | `cmd` | 执行远程命令 |
| `upload` | `local_path`, `remote_path` | 上传文件 |
| `upload_atomic` | `local_path`, `remote_path`, `temp_path` | 原子上传 |
| `download` | `local_path`, `remote_path` | 下载文件 |
| `mkdir` | `remote_path` | 创建目录 |
| `remove` | `remote_path` | 删除文件 |
| `chmod` | `remote_path`, `mode` | 修改权限 |
| `move` | `source`, `target` | 移动文件 |
| `sync` | `local_path`, `remote_path`, `direction` | 文件同步 |

`continue_on_error`: true 表示某步骤失败后继续执行后续步骤
