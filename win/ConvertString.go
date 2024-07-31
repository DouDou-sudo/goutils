package win

import (
	"log"
	"runtime"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// 解决windows终端下执行go程序输出中文时乱码的问题
func ConvertString(s string) (string, error) {
	if runtime.GOOS == "windows" { //如果是windows系统对string进行编码转换
		s, err := simplifiedchinese.GB18030.NewDecoder().String(s)
		if err != nil {
			log.Printf("编码转换失败:%v\n", err) //记录日志
			return "", err
		}
		return s, nil
	}
	return s, nil
}
