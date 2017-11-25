package main


/*
#define _XOPEN_SOURCE 600
#include <unistd.h>
#include <string.h>
#define _BSD_SOURCE
#include <termios.h>

void init(){
	int i = 0;
  	struct termios oldt, newt;
    tcgetattr( STDIN_FILENO, &oldt);
    newt = oldt;

    // newt.c_lflag &= ~( ICANON | ISIG);
	/////////////////
	cfmakeraw(&newt);
	// newt.c_lflag |= ECHO | ECHOE;
	newt.c_iflag |= ICRNL;
	newt.c_oflag |= OPOST;

	//printf("%x %x %x %x\n",newt.c_iflag,newt.c_oflag,newt.c_lflag,newt.c_cflag);
	//for(i=0;i< NCCS;i++){
	//	printf("%d ",newt.c_cc[i]);
	//}
	//printf("\n");
    tcsetattr( STDIN_FILENO, TCSANOW, &newt);
}

*/
import "C"
import (
	"fmt"
	"os"
	"net"
	"io"
)

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

func handleRead(conn *net.TCPConn, wt io.WriteCloser, done chan string) {
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

		// fmt.Println("recv :", string(buf[:l]))
		_, err = wt.Write(buf[:l])
		if err != nil {
			fmt.Println("Error to write message because of ", err)
			break
		}
	}

	// 忽略错误（尝试退出shell）
	wt.Close()
	done <- "Read"
}

func main(){
	C.init()

	conn, err := net.Dial("tcp", "192.168.1.219:40001")
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()

	done := make(chan string,2)
	defer close(done)

	tcpConn,_ := conn.(*net.TCPConn)

	go handleRead(tcpConn, os.Stdout, done)
	go handleWrite(tcpConn, os.Stdin, done)

	<- done
	<- done
}