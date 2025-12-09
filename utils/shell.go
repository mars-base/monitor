package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// ----------------- non realtime shell output, sync block -------------
// out, ok, code := utils.Shell("ls", "-l")
// utils.Log().Debug("exit code:", code)
// utils.Log().Debug("out:", out)
// if ok {
// 	for _, li := range utils.Split(out, "\n") {
// 		utils.Log().Debug(li)
// 	}
// } else {
// 	utils.Log().Error(out)
// }

// NOTE: if shell command has &&、|、； etc,
// need use like utils.Shell("bash", "-c", "ls -l && echo hello")

func Shell(cmdName string, args ...string) (string, bool, int) {
	var out string
	var ret bool
	var exitCode = 1 // exit code, 0 is success, non-zero is fail

	cmd := exec.Command(cmdName, args...) // 执行命令

	output, err := cmd.CombinedOutput()
	if err != nil {
		ret = false
		out = string(output)

		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Println(err.Error())
			exitCode = exitError.ExitCode()
		} else {
			out = err.Error()
		}
	} else {
		ret = true
		exitCode = 0
		out = string(output)
	}
	return out, ret, exitCode
}

// --------------------- realtime shell output --------------------
// cmd := "ping -c10 baidu.com"
// buf := make(chan string, 1)
// ctx := utils.ShellGetContextObject()
// go utils.ShellNonBlock(cmd, buf, ctx)
// for {
// 	line, ok := <-buf
// 	if !ok {
// 		break
// 	}
// 	if line == "done" {
// 		logger.Println("shell done.")
// 		break
// 	}
// 	logger.Printf("nonblock shell output: " + line)
// }

// if buf receives "done", then the shell output is done.
func ShellGetContextObject() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func(f context.CancelFunc) {
		time.Sleep(time.Second * 1)
	}(cancel)
	return ctx
}

func ShellNonBlock(cmd string, buf chan<- string, ctx context.Context) error {
	cc := exec.CommandContext(ctx, "bash", "-c", cmd)
	stdout, err := cc.StdoutPipe()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		reader := bufio.NewReader(stdout)
		for {
			select {
			case <-ctx.Done():
				buf <- "done"
				close(buf)
				return
			default:
				str, err := reader.ReadString('\n')
				if err != nil || err == io.EOF {
					buf <- "done"
					close(buf)
					return
				}
				buf <- str
			}
		}
	}(&wg)

	err = cc.Start()
	wg.Wait()

	return err
}
