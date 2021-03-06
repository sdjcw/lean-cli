package runtimes

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/aisk/chrysanthemum"
	"github.com/facebookgo/parseignore"
	"github.com/leancloud/lean-cli/lean/utils"
)

// defaultIgnorePatterns returns current runtime's default ignore patterns
func (runtime *Runtime) defaultIgnorePatterns() []string {
	switch runtime.Name {
	case "node.js":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			"node_modules/",
		}
	case "java":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			".project",
			".classpath",
			".settings/",
			"target/",
		}
	case "php":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			"vendor/",
		}
	case "python":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			"venv",
			"*.pyc",
			"__pycache__/",
		}
	default:
		panic("invalid runtime")
	}
}

func (runtime *Runtime) readIgnore(ignoreFilePath string) (parseignore.Matcher, error) {
	if ignoreFilePath == ".leanignore" && !utils.IsFileExists(filepath.Join(runtime.ProjectPath, ".leanignore")) {
		chrysanthemum.Println("> 没有找到 .leanignore 文件，根据项目文件创建默认的 .leanignore 文件")
		content := strings.Join(runtime.defaultIgnorePatterns(), "\r\n")
		err := ioutil.WriteFile(filepath.Join(runtime.ProjectPath, ".leanignore"), []byte(content), 0644)
		if err != nil {
			return nil, err
		}
	}

	content, err := ioutil.ReadFile(ignoreFilePath)
	if err != nil {
		return nil, err
	}

	matcher, errs := parseignore.CompilePatterns(content)
	if len(errs) != 0 {
		return nil, errs[0]
	}

	return matcher, nil
}
