# logger

## How to use

The duration for recording a single log file is 24 hours. After the duration exceeds 24 hours, a log file is automatically divided. All logs can be stored for a maximum of 7 x 24 hours.

### Def use

Logs are stored in the Log folder under the current program. The default log level is Info.

```go
logger.SetLoggerName("AppName")
logger.Info("haha")
// will save like AppName.log
```

### Change log level

```go
logger.SetLoggerName("AppName")
logger.Info("haha")
// [INFO]: 2022-02-11 08:51:16 - haha
logger.SetLoggerLevel(logrus.InfoLevel)
logger.Debug("haha")
// [DEBUG]: 2022-02-11 08:51:16 - haha
```

### Set log file save path

By default, will save at `./Logs/`

```go
logger.SetLoggerRootDir("/config/xxx")
// will save at /config/xxx/Logs
```

### Set logger name

By default, will save log files like: `logger.log`

```go
logger.SetLoggerName("AppName")
// will save like AppName.log
```

## Base on

* [sirupsen/logrus](https://github.com/sirupsen/logrus)
* [lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs)
* [t-tomalak/logrus-easy-formatter](https://github.com/t-tomalak/logrus-easy-formatter)