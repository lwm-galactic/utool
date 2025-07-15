package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/lwm-galactic/utool/pkg/app"
	"github.com/lwm-galactic/utool/pkg/log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	Python = "python"
	PIP    = "pip"
	Hasten = "https://pypi.tuna.tsinghua.edu.cn/simple"

	Configuration = "require.json"
)

var requireList = make(map[string][]string)

func init() {
	// 添加需求项
	requireList[Python] = []string{"pdf2docx", "you-get"}
}

func NewInitCommand() *app.Command {
	return app.NewCommand("init", "to init tool like download the dependence", app.WithCommandRunFunc(InitCommandRun))
}

func InitCommandRun(option app.CliOptions) error {
	log.Info("InitCommand call")
	for require, list := range requireList {
		switch require {
		case Python:
			return initPython(list)
		}
	}

	// log.Info("get cookie from web")
	log.Info("init success")
	return nil
}

func initPython(pkgs []string) error {
	log.Info("initPython call")
	installPkg, _ := loadPythonInstallPkg()
	log.Infof("need pkg: %v", pkgs)
	log.Infof("installPkg: %v", installPkg)

	pkgs = removeSliceElements(pkgs, installPkg)
	if len(pkgs) == 0 {
		log.Info("all packages already installed")
		return nil
	}
	for _, pkg := range pkgs {
		cmd := exec.Command(PIP, "install", pkg, "-i", Hasten)
		// 设置 stdout 和 stderr 的管道
		cmd.Stdout = os.Stdout // 直接输出到控制台
		cmd.Stderr = os.Stderr // 错误输出也直接输出到控制台

		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error init python for %v\n", err)
			return err
		}
		installPkg = append(installPkg, pkg)
	}
	install := make(map[string][]string)
	install[Python] = installPkg
	err := saveAllReadyInstall(install)
	if err != nil {
		return err
	}
	return nil
}

func removeSliceElements(a, b []string) []string {
	if b == nil || len(b) == 0 {
		return a
	}
	// 创建一个 map 用于快速查找
	exists := make(map[string]bool)
	for _, item := range b {
		exists[item] = true
	}

	// 遍历 a，只保留不在 b 中的元素
	var result []string
	for _, item := range a {
		if !exists[item] {
			result = append(result, item)
		}
	}

	return result
}
func loadPythonInstallPkg() ([]string, error) {
	homeDir, _ := os.UserHomeDir()
	folderPath := filepath.Join(homeDir, ".tool")
	log.Infof(folderPath)
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		log.Errorf("create work dir err: %v", err.Error())
		return nil, err
	}
	// 构建 require.json 路径
	jsonFilePath := filepath.Join(folderPath, Configuration)
	// 读取文件内容
	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		log.Errorf("读取 require.json 失败: %v", err)
		return nil, err
	}

	// 解析 JSON 数据
	var result map[string][]string
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Errorf("解析 require.json 失败: %v", err)
		return nil, err
	}

	return result[Python], nil
}
func saveAllReadyInstall(require map[string][]string) error {
	homeDir, _ := os.UserHomeDir()
	folderPath := filepath.Join(homeDir, ".tool")
	log.Infof(folderPath)
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		log.Errorf("create work dir err: %v", err.Error())
		return err
	}
	// 构建 require.json 路径
	jsonFilePath := filepath.Join(folderPath, Configuration)
	// 将 requireList 转换为 JSON 数据
	data, err := json.MarshalIndent(require, "", "    ")
	if err != nil {
		log.Fatalf("JSON 序列化失败: %v", err)
		return err
	}
	// 写入文件
	err = os.WriteFile(jsonFilePath, data, 0644)
	if err != nil {
		log.Fatalf("写入 require.json 失败: %v", err)
		return err
	}
	return nil
}
