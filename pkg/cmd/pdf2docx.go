package cmd

import (
	"fmt"
	"github.com/lwm-galactic/utool/pkg/cli"
	"os/exec"
	"path/filepath"

	"github.com/lwm-galactic/utool/pkg/app"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

const BaseCommandName = "pdf2docx"

const (
	Convert = "convert"
	GUI     = "gui"
)

type Pdf2DocxOptions struct {
	InputDir  string `mapstructure:"input-dir"`
	OutputDir string `mapstructure:"output-dir"`
	File      string `mapstructure:"file"`
	GUI       bool   `mapstructure:"gui"`
}

func (o *Pdf2DocxOptions) Validate() []error {
	var errs []error
	if o.InputDir != "" {
		if _, err := os.Stat(o.InputDir); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("input directory does not exist: %s", o.InputDir))
		}
	}
	if o.OutputDir != "" {
		if _, err := os.Stat(o.OutputDir); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("output directory does not exist: %s", o.OutputDir))
		}
	}
	if o.File != "" {
		// 检查是否以 .pdf 结尾
		if !strings.HasSuffix(strings.ToLower(o.File), ".pdf") {
			errs = append(errs, fmt.Errorf("file must have a .pdf extension"))
		}
	}
	return errs
}

func NewPdf2DocxOptions() *Pdf2DocxOptions {
	return &Pdf2DocxOptions{
		InputDir:  ".",
		OutputDir: ".",
		File:      "",
		GUI:       false,
	}
}

// AddFlags adds flags related to features for a specific api server to the
// specified FlagSet.
func (o *Pdf2DocxOptions) AddFlags(fs *pflag.FlagSet) {
	if fs == nil {
		return
	}

	fs.StringVar(&o.InputDir, "input-dir", o.InputDir,
		"设置 pdf2docx 要转换 pdf文件 所在的文件夹 (批量操作将该文件夹下的所有 *.pdf 文件转换成 *.docx)")

	fs.StringVar(&o.OutputDir, "output-dir", o.OutputDir,
		"设置 pdf2docx 转换 docx文件后 所在的文件夹")

	fs.StringVar(&o.File, "file", o.File, "指定一个pdf文件转换成docx")

	fs.BoolVar(&o.GUI, "gui", o.GUI, "是否使用图形化界面操作 true or false")
}

func (o *Pdf2DocxOptions) Flags() (fss cli.NamedFlagSets) {
	o.AddFlags(fss.FlagSet("pdf2docx"))
	return fss
}

func NewPdf2DocxCommand() *app.Command {
	return app.NewCommand("pdf2docx", "将 pdf 文件转换成 docx", app.WithCommandRunFunc(pdf2docxRun), app.WithCommandOptions(NewPdf2DocxOptions()))
}

func pdf2docxRun(option app.CliOptions) error {
	opts, ok := option.(*Pdf2DocxOptions)
	if !ok {
		return fmt.Errorf("pdf2docxRun option is invalid")
	}

	if opts.GUI {
		cmd := exec.Command(BaseCommandName, GUI)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("pdf2docxRun err: %v\n", err)
			return err
		}

		return nil
	}

	var pdfFiles []string
	var err error
	if opts.InputDir != "" {
		pdfFiles, err = readPdfFile(opts.InputDir)
		if err != nil {
			return err
		}
	}
	if opts.File != "" {
		pdfFiles, err = readPdfFile(opts.File)
	}

	for _, pdfFile := range pdfFiles {
		cmd := exec.Command(BaseCommandName, Convert, pdfFile,
			filepath.Join(opts.OutputDir, getFileName(pdfFile)))
		// 设置 stdout 和 stderr 的管道
		cmd.Stdout = os.Stdout // 直接输出到控制台
		cmd.Stderr = os.Stderr // 错误输出也直接输出到控制台

		fmt.Printf("Processing: %s\n", pdfFile)

		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing command for %s: %v\n", pdfFile, err)
			return err
		}
	}

	return nil
}

func getFileName(fileName string) string {
	return fmt.Sprintf("%s.docx", strings.TrimSuffix(filepath.Base(fileName), ".pdf"))
}

func readPdfFile(inputDir string) ([]string, error) {
	var pdfFiles []string
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是文件，并且以 .pdf 结尾
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".pdf") {
			pdfFiles = append(pdfFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking input directory:", err)
		return nil, err
	}
	return pdfFiles, nil
}
