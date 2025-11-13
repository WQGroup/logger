## 执行内容
- 扩展 Settings：加入 `MaxAgeDays`、`MaxSizeMB`，保留现有时间轮转与默认值。
- 新增 YAML 加载：`LoadSettingsFromYAML` 与 `SetLoggerFromYAML`，映射 `days_to_keep`、`level`、`max_size_mb`、`log_root`、`log_name_base`。
- 构建输出路径：使用 `time.Now().Format("2006/01/02")` 创建 `<root>/<YYYY>/<MM>/<DD>/`，月份与日两位数命名。
- 轮转选择：
  - `max_size_mb>0` → 使用 `lumberjack` 基于大小轮转（`MaxAgeDays`保留天数）。
  - 否则 → 使用 `rotatelogs` 时间轮转，文件模式为 `<root>/%Y/%m/%d/<base>--%H%M--.log`，`WithLinkName` 指向 `<root>/<base>.log`。
- 清理空目录：实现 `CleanupExpiredLogs(root string, days int)`，删除过期日目录并递归移除空的月/年目录；初始化时运行一次并设置轻量定时清理。
- 保持 API 兼容与 Windows GUI 分支逻辑。
- 添加与运行测试，`go build -v` 验证。

## 影响与兼容
- 对外调用不变；新增 YAML 入口与分层存储。
- `CurrentFileName()` 在大小轮转下返回当前活跃文件路径，在时间轮转下保持已有行为。

现在开始实施并验证。