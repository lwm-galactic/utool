package log

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

const (
	flagLevel             = "log.level"
	flagDisableCaller     = "log.disable-caller"
	flagDisableStacktrace = "log.disable-stacktrace"
	flagFormat            = "log.format"
	flagOutputPaths       = "log.output-paths"
	flagDevelopment       = "log.development"
	flagName              = "log.name"
	// flagMaxBackups        = "log.max-backups"
	// flagMaxAge            = "log.max-age"
	// flagMaxSize           = "log.max-size"
	// flagRotateInterval    = "log.rotate-interval"
	flagErrorOutputPaths = "log.error-output-paths"

	consoleFormat = "console"
	jsonFormat    = "json"
)

// Options 日志配置项.
type Options struct {
	OutputPaths       []string `json:"output-paths"       mapstructure:"output-paths"`       // 输出位置，例如 ["stdout", "/var/log/app.log"]
	Level             string   `json:"level"              mapstructure:"level"`              // 日志级别 debug/info/warn/error
	Format            string   `json:"format"             mapstructure:"format"`             // 格式 json/console
	DisableCaller     bool     `json:"enable-call"        mapstructure:"disable-call"`       // 是否启用 call
	DisableStacktrace bool     `json:"disable-stacktrace" mapstructure:"disable-stacktrace"` // 是否记录 error 的 stack trace
	Development       bool     `json:"development"        mapstructure:"development"`        // 是否 DPanic
	ErrorOutputPaths  []string `json:"error-output-paths" mapstructure:"error-output-paths"` // 错误日志输出途径

	// MaxSize        int           `json:"max-size"           mapstructure:"max-size"`        // 文件最大 MB(如果用了 lumberjack)
	// MaxBackups     int           `json:"max-backups"        mapstructure:"max-backups"`     // 最大保留旧文件数
	// MaxAge         time.Duration `json:"max-age"            mapstructure:"max-age"`         // 日志保留时间
	// RotateInterval time.Duration `json:"rotate-interval"    mapstructure:"rotate-interval"` // 轮转间隔（如 24h）

	Name string `json:"name"               mapstructure:"name"` // server Name

	// EnableColor bool `json:"enable-color"       mapstructure:"enable-color"`
}

// Validate 验证配置是否符合规范.
func (o *Options) Validate() []error {
	var errs []error

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(o.Level)); err != nil {
		errs = append(errs, err)
	}

	format := strings.ToLower(o.Format)
	if format != consoleFormat && format != jsonFormat {
		errs = append(errs, fmt.Errorf("not a valid log format: %q", o.Format))
	}

	return errs
}

// AddFlags 构建.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&o.ErrorOutputPaths, flagErrorOutputPaths, o.ErrorOutputPaths, "Error output paths of log.")
	// fs.BoolVar(&o.EnableColor, flagEnableColor, o.EnableColor, "Enable output ansi colors in plain format logs.")

	fs.StringVar(&o.Level, flagLevel, o.Level, "Minimum log output `LEVEL`.")
	fs.BoolVar(&o.DisableCaller, flagDisableCaller, o.DisableCaller, "Disable output of caller information in the log.")
	fs.BoolVar(&o.DisableStacktrace, flagDisableStacktrace,
		o.DisableStacktrace, "Disable the log to record a stack trace for all messages at or above panic level.")
	fs.StringVar(&o.Format, flagFormat, o.Format, "Log output `FORMAT`, support plain or json format.")
	fs.StringSliceVar(&o.OutputPaths, flagOutputPaths, o.OutputPaths, "Output paths of log.")
	fs.BoolVar(
		&o.Development,
		flagDevelopment,
		o.Development,
		"Development puts the logger in development mode, which changes "+
			"the behavior of DPanicLevel and takes stacktraces more liberally.",
	)
	fs.StringVar(&o.Name, flagName, o.Name, "The name of the logger.")
	// fs.IntVar(&o.MaxSize, flagMaxSize, o.MaxSize, "Maximum size of the logger.")
	// fs.IntVar(&o.MaxBackups, flagMaxBackups, o.MaxBackups, "Maximum backups of the logger.")
	// fs.DurationVar(&o.MaxAge, flagMaxAge, o.MaxAge, "Maximum age of the logger.")
	// fs.DurationVar(&o.RotateInterval, flagRotateInterval, o.RotateInterval, "Rotate interval of log entries.")
}

// NewOptions 创建一个默认的配置项.
func NewOptions() *Options {
	return &Options{
		Level:             zapcore.InfoLevel.String(),
		DisableCaller:     false,
		DisableStacktrace: false,
		Format:            consoleFormat,
		Development:       false,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},

		// MaxBackups:     1000,
		// MaxAge:         24 * 30 * time.Hour,
		// MaxSize:        1024 * 1024 * 1024,
		// RotateInterval: 24 * time.Hour,
	}
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)
	return string(data)
}

// Build constructs a global zap logger from the Config and Options.
func (o *Options) Build() error {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(o.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	encodeLevel := zapcore.CapitalLevelEncoder
	if o.Format == consoleFormat {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zc := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       o.Development,
		DisableCaller:     o.DisableCaller,
		DisableStacktrace: o.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: o.Format,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "timestamp",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    encodeLevel,
			EncodeTime:     timeEncoder,
			EncodeDuration: milliSecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      o.OutputPaths,
		ErrorOutputPaths: o.ErrorOutputPaths,
	}
	logger, err := zc.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		return err
	}
	zap.RedirectStdLog(logger.Named(o.Name))
	zap.ReplaceGlobals(logger)

	return nil
}
