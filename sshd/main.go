package main

/*
#define _XOPEN_SOURCE 600
#include <stdlib.h>
#include <stdio.h>
#include <fcntl.h>
#include <errno.h>
#include <unistd.h>
#define _BSD_SOURCE
#include <termios.h>

int fd1(){
	int fdm;
	int rc;
	fdm = posix_openpt(O_RDWR);
	if (fdm < 0)
	{
		fprintf(stderr, "Error %d on posix_openpt()\n", errno);
		return -1;
	}

	rc = grantpt(fdm);
	if (rc != 0)
	{
		fprintf(stderr, "Error %d on grantpt()\n", errno);
		return -1;
	}

	rc = unlockpt(fdm);
	if (rc != 0)
	{
		fprintf(stderr, "Error %d on unlockpt()\n", errno);
		return -1;
	}
	return fdm;
}

int fd2(int fdm) {
	int fds = -1;
	int rc = -1;
	struct termios slave_orig_term_settings; // Saved terminal settings
	struct termios new_term_settings; // Current terminal settings

	// Open the slave side ot the PTY
	fds = open(ptsname(fdm), O_RDWR);
	if (fds < 0)
	{
		fprintf(stderr, "Error %d on posix_openpt()\n", errno);
		return -1;
	}

	// Save the defaults parameters of the slave side of the PTY
	rc = tcgetattr(fds, &slave_orig_term_settings);

	// Set RAW mode on slave side of PTY
	new_term_settings = slave_orig_term_settings;
	cfmakeraw (&new_term_settings);
	new_term_settings.c_lflag |= ECHO | ECHOE;
	new_term_settings.c_iflag |= ICRNL;
	new_term_settings.c_oflag |= OPOST;
	tcsetattr (fds, TCSANOW, &new_term_settings);

	return fds;
}

*/
import "C"
import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting the server ...")
	// 创建 listener
	listener, err := net.Listen("tcp", "0.0.0.0:40001")
	if err != nil {
		fmt.Println("Error listening", err.Error())
		return //终止程序
	}

	// 统计和管理连接
	// 增加和统计连接
	newConnChan := make(chan *net.TCPConn)
	defer close(newConnChan)
	// 回收连接
	destoryConnChan := make(chan *net.TCPConn)
	defer close(destoryConnChan)
	// 退出整个程序
	doneChan := make(chan bool, 1)
	defer close(doneChan)
	// 信号
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go doManageStuff(newConnChan, destoryConnChan, doneChan, sigs, listener)

	// 监听并接受来自客户端的连接
	for {
		conn, err := listener.Accept()
		if conn == nil {
			fmt.Println("listener accept ended", err)
			break
		}
		if err != nil {
			fmt.Println("Error accepting", err.Error())
			break
		}

		tcpConn, _ := conn.(*net.TCPConn)
		// 添加链接
		newConnChan <- tcpConn
		go doServerStuff(tcpConn, destoryConnChan)
	}

	<-doneChan
	fmt.Println("main loop exit")
}

func doManageStuff(in, out chan *net.TCPConn, doneChan chan bool, sig chan os.Signal, listener net.Listener) {
	breakFlag := false
	connCount := 0
	conns := make(map[*net.TCPConn]bool)
	for {
		if breakFlag && connCount == 0 {
			break
		}

		select {
		case conn := <-in:
			conns[conn] = false
			connCount += 1
			fmt.Println("new connection", connCount, conn)
		case conn := <-out:
			if _, ok := conns[conn]; ok {
				delete(conns, conn)
				connCount -= 1

				fmt.Println("close connection", conn)
				if err := conn.Close(); err != nil {
					fmt.Println(err)
				}
			}
			if breakFlag && connCount == 0 {
				fmt.Println("break manage", connCount, conn)
				break
			}
		case <-sig:
			if breakFlag {
				continue
			}

			breakFlag = true
			if err := listener.Close(); err != nil {
				fmt.Println(err)
			}
			if connCount == 0 {
				break
			} else {
				for k, _ := range conns {
					if err := k.CloseWrite(); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}

	fmt.Println("manager exit")
	doneChan <- true
}

func handleWrite(conn *net.TCPConn, reader *os.File, done chan string) {
	// reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 1024)
	for {
		l, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("reader from tty end")
			} else {
				fmt.Println("Error to read message because of ", err)
			}
			break
		}

		_, err = conn.Write(buf[:l])
		if err != nil {
			fmt.Println("Error to write message because of ", err)
			break
		}
	}
	// 尝试关闭tcp连接
	conn.CloseWrite()
	done <- "Sent"
}

func handleRead(conn *net.TCPConn, wt *os.File, done chan string) {
	buf := make([]byte, 1024)
	for {
		l, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("tcp read end")
			} else {
				fmt.Println("Error to read message because of ", err)
			}
			break
		}

		_, err = wt.Write(buf[:l])
		if err != nil {
			fmt.Println("Error to write message because of ", err)
			break
		}
	}

	// 忽略错误（尝试退出shell）
	wt.Write([]byte("exit\n"))
	done <- "Read"
}

func doServerStuff(conn *net.TCPConn, out chan *net.TCPConn) {
	/////////////////////////////////////////////////////////
	cmd := exec.Command("/bin/bash")
	//cmd := exec.Command("/bin/login")

	fd := C.fd1()
	fd2 := C.fd2(fd)
	fmt.Println("tty ", fd, fd2)
	pty1 := os.NewFile(uintptr(fd), "ptsm")
	defer pty1.Close()

	tty := os.NewFile(uintptr(fd2), "pty")

	cmd.Stdout = tty
	cmd.Stdin = tty
	cmd.Stderr = tty
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true
	if err := cmd.Start(); err != nil {
		fmt.Printf("err %s %v", "main end", err)
		tty.Close()
		out <- conn
		return
	}
	tty.Close()

	done := make(chan string)
	defer close(done)

	go handleWrite(conn, pty1, done)
	go handleRead(conn, pty1, done)

	// 等待go结束
	<-done
	<-done
	// Wait后台进程
	cmd.Wait()

	// 销毁tcp连接
	out <- conn
}
