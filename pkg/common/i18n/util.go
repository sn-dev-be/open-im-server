package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

// convert yaml Bundle to json jsonData mapping
func yamlToJson(buf []byte) (any, error) {
	j, err := yaml.YAMLToJSON(buf)
	if err != nil {
		return nil, err
	}

	m := make(map[string]any)
	if err = json.Unmarshal(j, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func getRootPath() (string, error) {
	// 获取可执行文件的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// 使用 filepath 包获取可执行文件所在目录
	exeDir := filepath.Dir(exePath)

	// 从可执行文件所在目录向上遍历，直到找到包含 go.mod 文件的目录，作为项目的根路径
	rootPath := exeDir
	for {
		if _, err := os.Stat(filepath.Join(rootPath, "go.mod")); err == nil {
			return rootPath, nil
		}

		// 向上一级目录
		parentDir := filepath.Dir(rootPath)
		if parentDir == rootPath {
			// 到达文件系统根目录，仍未找到 go.mod 文件
			return "", fmt.Errorf("go.mod not found in project")
		}

		rootPath = parentDir
	}
}
