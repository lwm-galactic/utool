package cmd

import (
	"fmt"
	"github.com/lwm-galactic/utool/pkg/app"
	"github.com/lwm-galactic/utool/pkg/cli"
	"github.com/lwm-galactic/utool/pkg/log"
	"github.com/spf13/pflag"
	"os"
	"os/exec"
	"strings"
)

type DownloadOptions struct {
	Url       string `mapstructure:"url"`
	File      string `mapstructure:"file"`
	OutputDir string `mapstructure:"output-dir"`
}

func (o *DownloadOptions) Validate() []error {
	var errs []error
	if o.Url == "" && o.File == "" {
		errs = append(errs, fmt.Errorf("url or file is required"))
	}
	return errs
}

// AddFlags adds flags related to features for a specific api server to the
// specified FlagSet.
func (o *DownloadOptions) AddFlags(fs *pflag.FlagSet) {
	if fs == nil {
		return
	}

	fs.StringVar(&o.OutputDir, "output-dir", o.OutputDir,
		"设置下载输出文件夹")

	fs.StringVar(&o.File, "file", o.File, "指定一个文件,一行是一个下载地址 用于批量下载")

	fs.StringVar(&o.Url, "url", o.Url, "指定一个下载地址")
}

func (o *DownloadOptions) Flags() (fss cli.NamedFlagSets) {
	o.AddFlags(fss.FlagSet("download"))
	return fss
}

func NewDownloadCommand() *app.Command {
	return app.NewCommand("download", "用于下载 bilibili | youtube 的视频工具", app.WithCommandRunFunc(downloadRun), app.WithCommandOptions(NewDownloadOptions()))
}
func NewDownloadOptions() *DownloadOptions {
	return &DownloadOptions{
		OutputDir: ".",
	}
}

func downloadRun(option app.CliOptions) error {
	opts, ok := option.(*DownloadOptions)
	if !ok {
		return fmt.Errorf("downloadRun: invalid options")
	}

	// 确保输出目录存在
	if opts.OutputDir != "" {
		err := os.MkdirAll(opts.OutputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}
	}

	// 构建 you-get 命令参数
	cmdArgs := []string{}
	if opts.OutputDir != "" {
		cmdArgs = append(cmdArgs, "-o", opts.OutputDir)
	}

	// 判断是通过 URL 下载还是通过文件下载
	if opts.Url != "" {
		cmdArgs = append(cmdArgs, opts.Url)
	} else if opts.File != "" {
		cmdArgs = append(cmdArgs, "-i", opts.File)
	} else {
		return fmt.Errorf("必须指定 url 或 file 参数")
	}

	// 打印正在执行的命令
	log.Infof("执行命令: you-get %s", strings.Join(cmdArgs, " "))

	// 创建命令
	cmd := exec.Command("you-get", cmdArgs...)

	// 设置标准输出和错误输出流（实时输出）
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行命令
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("you-get 执行失败: %v", err)
	}

	return nil
}
