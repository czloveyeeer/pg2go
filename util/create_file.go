package util

import (
	"fmt"
	"os"
	"os/exec"
)

//创建文件
func CreateFile(_tableName, s string, outDir string) error {
	exist, err := pathExists(outDir)
	if err != nil {
		return err
	}
	//文件夹不存在 创建
	if !exist {
		fmt.Printf("\"%s\" 输出目录不存在，创建目录...\n", outDir)
		err := os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(PathTrim(fmt.Sprintf("%s/%s.go", outDir, _tableName)))
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)
	if err != nil {
		fmt.Printf("错误! 创建 %s.go 文件失败，err:%v", _tableName, err.Error())
		return err
	} else {
		_, err = f.Write([]byte(s))
		if err != nil {
			fmt.Printf("错误! 创建 %s.go 文件失败，err:%v", _tableName, err.Error())
			return err
		}
	}
	fmt.Printf("创建 %s.go 文件成功，路径为：%s\n", _tableName, f.Name())
	fmtFile(f.Name())
	return nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func fmtFile(filename string) {
	command := exec.Command("gofmt", "-w", filename)
	fmt.Println(filename)
	out, err := command.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:\n%s\n", string(out))
	}
	fmt.Printf("combined out:\n%s\n", string(out))
}
