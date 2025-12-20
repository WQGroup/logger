package logger

import (
	"sync"
	"testing"
	"github.com/sirupsen/logrus"
	"github.com/lestrrat-go/file-rotatelogs"
)

// testStateBackup 安全地备份和恢复logger状态
// 用于测试中避免竞态条件
type testStateBackup struct {
	loggerBase          *logrus.Logger
	rotateLogsWriter    *rotatelogs.RotateLogs
	currentLogFileFPath string
	// 移除: loggerOnce sync.Once - sync.Once不应该被复制，应重新初始化
}

// backupState 安全备份当前状态
// 必须在持有锁的情况下调用
func backupState() *testStateBackup {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	return &testStateBackup{
		loggerBase:          loggerBase,
		rotateLogsWriter:    rotateLogsWriter,
		currentLogFileFPath: currentLogFileFPath,
	}
}

// restoreState 安全恢复状态
// 在defer中使用，确保状态恢复的原子性
func (b *testStateBackup) restoreState() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	loggerBase = b.loggerBase
	rotateLogsWriter = b.rotateLogsWriter
	currentLogFileFPath = b.currentLogFileFPath
	// 重置 loggerOnly 标记
	loggerOnce = sync.Once{}
}

// resetState 完全重置 logger 状态
// 用于需要完全清理的测试
func resetState() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	_ = closeOldResources() // 忽略错误，因为这是测试清理代码
	loggerBase = nil
	rotateLogsWriter = nil
	currentLogFileFPath = ""
	loggerOnce = sync.Once{}
}

// withBackup 提供标准的测试模式
// 确保测试隔离和状态恢复
func withBackup(t *testing.T, testFunc func()) {
	backup := backupState()
	defer backup.restoreState()
	testFunc()
}