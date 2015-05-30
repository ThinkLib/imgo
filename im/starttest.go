package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"encoding/json"
	"time"
)

type IMResponse struct {
	Status int                               `json:"status"` //状态 0成功，非0错误
	Msg    string                            `json:"msg"`    //消息
	Data   map[string]map[string]interface{} `json:"data"`   //数据
	Refer  string                            `json:"refer"`  //来源
}

func main() {
	var recv chan string = make(chan string)

	for i := 0; i<150; i++ {
		go testTcp()
	}

	<- recv
}

// 测试长连接数量
func testConn() {
	_, err := net.Dial("tcp", "127.0.0.1:9090")

	if err != nil {
		log.Fatal(err)
	}
}

// 测试tcp转发
func testTcp() {
	conn, err := net.Dial("tcp", "127.0.0.1:9090")

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	var recv chan string = make(chan string)

	go func() {
		for {
			if line, _, err := reader.ReadLine(); err == nil {
				var str string
				str = string(line)
				log.Println(str)
				recv <- str
			} else {
				os.Exit(0)
			}
		}
	}()

	var res IMResponse
	for {
		// 收到消息，然后再回复
		line := <-recv
		if line != "" {
			json.Unmarshal([]byte(line), &res)
			switch res.Refer {
			case "TCP_TEST":
				// 发送消息
				line = "{\"command\":\"TEST_TCP\",\"data\":\"Hello.\"}"

				time.Sleep(10 * time.Second)
			}
		}
		writer.WriteString(string(line) + "\n")
		err := writer.Flush()
		if (err != nil) {
			os.Exit(0)
		}
	}
}

// 测试转发，以及数据库能力
func test(sender string, token string, receiver string) {
	conn, err := net.Dial("tcp", "127.0.0.1:9090")

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	var recv chan string = make(chan string)

	go func() {
		for {
			if line, _, err := reader.ReadLine(); err == nil {
				var str string
				str = string(line)
				log.Println(str)
				recv <- str
			} else {
				os.Exit(0)
			}
		}
	}()

	var res IMResponse
	var ticket string
	for {
		// 收到消息，然后再回复
		line := <-recv
		if line != "" {
			json.Unmarshal([]byte(line), &res)
			switch res.Refer {
			case "GET_KEY_RETURN":
				// 建立连接
				line = "{\"command\":\"GET_CONN\",\"data\":{\"user\":{\"id\":\"" + sender + "\",\"token\":\"" + token + "\",\"key\":\"" + res.Data["conn"]["key"].(string) + "\"}}}"
			case "GET_CONN_RETURN":
				// 创建会话
				line = "{\"command\":\"CREATE_SESSION\",\"data\":{\"session\":{\"sender\":\"" + sender + "\",\"receiver\":\"" + receiver + "\",\"token\":\"" + token + "\"}}}"

			case "CREATE_SESSION_RETURN":
				// 发送消息
				ticket = res.Data["session"]["ticket"].(string)
				line = "{\"command\":\"SEND_MSG\",\"data\":{\"message\":{\"content\":\"Hello  World\",\"ticket\":\"" + ticket + "\",\"token\":\"" + token + "\"}}}"

			case "PUSH_MSG":
				// 发送消息
				line = "{\"command\":\"SEND_MSG\",\"data\":{\"message\":{\"content\":\"Hello  World\",\"ticket\":\"" + res.Data["session"]["ticket"].(string) + "\",\"token\":\"" + token + "\"}}}"

				time.Sleep(5 * time.Second)
			}
		}
		writer.WriteString(string(line) + "\n")
		err := writer.Flush()
		if (err != nil) {
			os.Exit(0)
		}
	}
}