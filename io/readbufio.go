package io

import (
	"bufio"
	"context"
	"io"
	"log"
)

type printBuffer func(string) error

// 按行使用bufio.NewScanner读取,使用printBuffer进行数据处理,读取过程可被终止
func BufioLineContext(ctx context.Context, r io.Reader, printBuffer printBuffer) error {
	readscanner := bufio.NewScanner(r)
	for {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err()) // 日志记录错误
			return ctx.Err()
		default:
			for readscanner.Scan() { // scanner.Text() 返回读取到的行(不包含行尾的\n),如果出现错误或者到达末尾都会返回false
				if err := printBuffer(readscanner.Text() + "\n"); err != nil { //数据处理
					return err
				}
				// fmt.Println(readscanner.Text()) //直接打印
			}
			if err := readscanner.Err(); err != nil {
				log.Printf("read io failed:%v\n", err) // 日志记录错误
				return err
			}
			return nil
		}
	}
}

// 按行使用bufio.NewScanner读取,使用printBuffer进行数据处理
func BufioLine(r io.Reader, printBuffer printBuffer) error {
	readscanner := bufio.NewScanner(r)
	for readscanner.Scan() { // scanner.Text() 返回读取到的行(不包含行尾的\n),如果出现错误或者到达末尾都会返回false
		if err := printBuffer(readscanner.Text() + "\n"); err != nil { //数据处理
			return err
		}
		// fmt.Println(readscanner.Text()) //直接打印
	}
	if err := readscanner.Err(); err != nil {
		log.Printf("read io failed:%v\n", err) // 日志记录错误
		return err
	}
	return nil
}

// 按lenbyte个byte大小依次使用bufio读取,使用printBuffer进行数据处理
func BufioByte(r io.Reader, lenbyte int, printBuffer printBuffer) error {
	bufioread := bufio.NewReader(r)
	buf := make([]byte, 0, lenbyte) //默认1kB缓冲区大小当输入有汉字时,必须为4的倍数,否则会乱码
	for {
		n, err := bufioread.Read(buf[len(buf):cap(buf)])
		if err != nil { //有err进行判断,不是EOF则返回err
			if err == io.EOF { //如果是EOF则判断n是否为0,如果n为0则返回nil
				if n != 0 { //如果n!=0则进行打印再返回nil
					if err := printBuffer(string(buf[:n])); err != nil {
						return err
					}
					// fmt.Print(string(buf[:n])) //直接打印
				}
				return nil
			}
			log.Printf("read io failed:%v\n", err) // 日志记录错误
			return err
		} //如果没err打印
		if err := printBuffer(string(buf[:n])); err != nil {
			return err
		}
		// fmt.Print(string(buf[:n])) //直接打印
	}
}

// 按分隔符delim分割,再使用bufio读取,使用printBuffer进行数据处理
func BufioDelim(r io.Reader, delim byte, printBuffer printBuffer) error {
	readbufio := bufio.NewReader(r)
	for {
		line, err := readbufio.ReadString(delim) // 不去除delim,也可以使用linebytes, err := r.ReadBytes(delim)返回的结果为[]byte
		if err != nil {
			if err == io.EOF {
				if line != "" { //如果line!=0则进行打印再返回nil
					if err := printBuffer(line); err != nil {
						return err
					}
					// fmt.Print(line) //直接打印
				}
				return nil //如果err是EOF则返回nil
			}
			log.Printf("read io failed:%v\n", err) // 日志记录错误return
			return err
		}
		if err := printBuffer(line); err != nil {
			return err
		}
		// fmt.Print(line) //直接打印
	}
}
