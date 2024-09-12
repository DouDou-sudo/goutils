package exec

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"sync"

	"github.com/Wybal/goutils/win"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func cmd() {
	// 0、在PATH中检测命令是否存在,如果存在返回目录的绝对路径,不存在返回not found
	// cmdpath, err := exec.LookPath("ls1")
	// fmt.Println(cmdpath)
	// 对命令执行的err进行判断
	// if err != nil {
	// 	log.Fatalf("cmd.run failed with %s\n", err)
	// }

	// cmd := "   ls -l  "
	// // cmd = strings.TrimSpace(cmd) // 当cmd前后有空格时,输出为空命令没有执行,所有先去掉前后的空格,使用bash -c cmd没有这个问题
	// execcmd := exec.Command("bash", "-c", cmd)

	// 1、只执行命令,不获取命令返回的标准输出和错误输出
	// cmd := exec.Command("ls")
	// _ = cmd.Run()

	// 一、使用变量对输出进行保存,再从变量获取输出,等待命令执行完后再对变量进行读取
	// 1、执行命令并获取标准输出至第一个参数out,不获取错误输出,第二个个参数为cmd.run()时的err
	// out, _ := exec.Command("ls", "-l").Output()
	// fmt.Print(string(out))

	// 2、执行命令并获取标准输出和错误输出合并返回给第一个参数out,第二个个参数为cmd.run()时的err
	// out, _ := exec.Command("ls", "-l").CombinedOutput()
	// fmt.Print(string(out))

	// 3、执行命令并分别获取标准输出和错误输出到不同的变量
	// cmd := exec.Command("ls", "-l")
	// var stdout, stderr bytes.Buffer		//var b bytes.Buffer	//和exec.Command("").CombinedOutput()同理
	// cmd.Stdout = &stdout                 //cmd.Stdout = &b
	// cmd.Stderr = &stderr					//cmd.Stderr = &b
	// // 直接将命令的标准输出和错误输出重定向到os的标准输出和错误输出
	// // cmd.Stderr = os.Stderr
	// // cmd.Stdout = os.Stdout
	// err := cmd.Run()
	// fmt.Printf("out:%s\nerr:%s\n", stdout.String(), stderr.String())

	// 二、使用管道,直接使用如下方法1即可
	// 1、使用bash -c 加复杂命令
	// cmd := "cat /proc/cpuinfo | egrep '^model name' | uniq | awk  '{print substr($0, index($0,$4))}'"
	// out, err := exec.Command("bash", "-c", cmd).Output()
	// // out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	// fmt.Print(string(out))
	// fmt.Print(err)
	// 输出
	// Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
	// <nil>

	// 2、使用StdoutPipe
	// cmd1 := exec.Command("bash", "-c", "ls")
	// cmd2 := exec.Command("bash", "-c", "wc", "-l")
	// // 将cmd1的标准输出管道指向cmd2的标准输入就实现了类似linux的 | 功能
	// cmd2.Stdin, _ = cmd1.StdoutPipe()
	// var out bytes.Buffer
	// cmd2.Stdout = &out
	// // 执行命令不等待完成
	// _ = cmd2.Start()
	// // cmd.Run()执行命令阻塞等待完成
	// _ = cmd1.Start()
	// // 和Start()配合使用，cmd必须由Start()启动。等待Start()的stdin,stdout,stderr都完成，Wait()会释放cmd相关的资源
	// _ = cmd1.Wait()
	// _ = cmd2.Wait()
	// fmt.Println(out.String())
	// cmd := "ping www.baidu.com -t"
	// Command(cmd)
	// f, _ := os.Open("test.txt")
	// bufioline(f)
	// bufiobyte(f)
	// bufiodelim(f, '\n')
}

func main() {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// go func() {
	// 	time.Sleep(3 * time.Second)
	// 	cancel()
	// }()
	fmt.Println(Command("ping www.baidu.com -t"))
	// os.ReadFile()
	// f, _ := os.Open("test.txt")
	// defer f.Close()
	// bufiobyte(f)
	// bufioline(f)
	// bufiodelim(f, '\n')
}

// 三、实时显示命令输出
// 当没有使用管道时,先获取输出再对输出进行读取,需要2步且读取输出不实时.
// 当命令的输出量特别大,而没有尽快读取输出,那么命令的输出缓冲区可能会填满,导致命令挂起,同时也会阻塞cmd.Wait()去释放资源
func Command(cmdName string) error {
	cmd := exec.Command("bash", "-c", cmdName)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("get %v stdoutPipe err:%v\n", cmd, err)
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("get %v stderrPipe err:%v\n", cmd, err)
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { //在commandContext里启动两个协程,分别读取stdout和stderr
		defer wg.Done()
		// Bufiodelim(stdoutPipe, '\n')
		// Bufioline(stdoutPipe)
		Bufiobyte(stdoutPipe)
	}()
	go func() {
		defer wg.Done()
		// Bufiodelim(stderrPipe, '\n')
		// Bufioline(stderrPipe)
		Bufiobyte(stderrPipe)
	}()
	if err := cmd.Start(); err != nil { // 执行命令
		return err
	}
	if err := cmd.Wait(); err != nil { // 等待命令执行完成
		return err
	}
	wg.Wait() //等待协程执行完毕
	return nil
}

// 四、实时显示命令输出，并可以终止命令
func CommandContext(ctx context.Context, cmdName string) error {
	cmd := exec.CommandContext(ctx, "bash", "-c", cmdName)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("get %v stdoutPipe err:%v\n", cmd, err)
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("get %v stderrPipe err:%v\n", cmd, err)
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { //在commandContext里启动两个协程,分别读取stdout和stderr
		defer wg.Done()
		BufioLineContext(ctx, stdoutPipe)
	}()
	go func() {
		defer wg.Done()
		BufioLineContext(ctx, stderrPipe)
	}()
	if err := cmd.Start(); err != nil { // 执行命令
		return err
	}
	if err := cmd.Wait(); err != nil { // 等待命令执行完成
		return err
	}
	wg.Wait() //等待协程执行完毕
	return nil
}

// 按行读取+可终止
func BufioLineContext(ctx context.Context, r io.Reader) error {
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

// 按行使用bufio.NewScanner读取
func Bufioline(r io.Reader) error {
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

// 按1024个字节大小依次使用bufio读取，对于可能没有换行的大文件
func Bufiobyte(r io.Reader) error {
	bufioread := bufio.NewReader(r)
	buf := make([]byte, 0, 1024) //默认1kB缓冲区大小当输入有汉字时，必须为4的倍数，否则会乱码
	for {
		n, err := bufioread.Read(buf[len(buf):cap(buf)])
		if err != nil { //有err进行判断，不是EOF则返回err
			if err == io.EOF { //如果是EOF则判断n是否为0，如果n为0则返回nil
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

// 按自定义分隔符分割再依次使用bufio读取
func Bufiodelim(r io.Reader, delim byte) error {
	readbufio := bufio.NewReader(r)
	for {
		line, err := readbufio.ReadString(delim) // 不去除delim，也可以使用linebytes, err := r.ReadBytes(delim)返回的结果为[]byte
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

// 将数据的处理逻辑抽离出来
func printBuffer(s string) error {
	s, err := win.ConvertString(s)
	if err != nil {
		return err
	}
	fmt.Print(s)
	return nil
}

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
