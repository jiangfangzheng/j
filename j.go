package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

func main() {
	cmdType := 0
	var buffer bytes.Buffer
	for i, v := range os.Args {
		if i == 0 {
			continue
		}
		if i == 1 && v == "cnv" {
			cmdType = 1
			buffer.WriteString(os.Args[i+1])
			break
		} else {
			cmdType = 0
			buffer.WriteString(" ")
			buffer.WriteString(v)
		}
	}
	if cmdType == 0 {
		execCmd(buffer.String())
	} else if cmdType == 1 {
		cmdStr := fmt.Sprintf("ffmpeg -i %s -acodec copy -vcodec copy -f mp4 %s.mp4",
			buffer.String(), buffer.String())
		execCmd(cmdStr)
	}
}

func execCmd(cmdStr string) {
	startT := time.Now()
	fmt.Printf("~exec cmd:%s\n", cmdStr)
	ctx, cancel := context.WithCancel(context.Background())
	go func(cancelFunc context.CancelFunc) {
		time.Sleep(3 * time.Second)
		cancelFunc()
	}(cancel)
	// 执行命令, 命令不会结束
	Command(ctx, cmdStr)
	fmt.Printf("~exec cmd: end cost:%v\n", time.Since(startT))
}

func read(ctx context.Context, wg *sync.WaitGroup, std io.ReadCloser) {
	reader := bufio.NewReader(std)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			byte2String := ConvertByte2String([]byte(readString), "GB18030")
			fmt.Print(byte2String)
		}
	}
}

func Command(ctx context.Context, cmd string) error {
	c := exec.Command("cmd", "/C", cmd) // windows
	//c := exec.Command("bash", "-c", cmd)  // mac or linux
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	// 因为有2个任务, 一个需要读取stderr 另一个需要读取stdout
	wg.Add(2)
	go read(ctx, &wg, stderr)
	go read(ctx, &wg, stdout)
	// 这里一定要用start,而不是run 详情请看下面的图
	err = c.Start()
	// 等待任务结束
	wg.Wait()
	return err
}

func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}
