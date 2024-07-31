package main

import (
	"bufio"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"time"
)

var done chan interface{}

// 处理接收请求
func receiveHandler(connection *websocket.Conn) {
	close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

func main() {
	done = make(chan interface{})

	//服务器地址 websocket 统一使用 ws://
	url := "ws://192.168.200.199:8200/stt/streaming"
	//使用默认拨号器，向服务器发送连接请求
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{
		"Authorization": []string{"Bearer eyJraWQiOiJwc3NvLWtleS1pZCIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiJkZW1vLXVzZXIiLCJhdWQiOiJEUFVGUEpIQ1NJIiwibmJmIjoxNzIyMzE4MjQzLCJzY29wZSI6WyJhcHR0cy5kaWN0Iiwib3BlbmlkIiwiYXBhc3IucnRzciIsImFwdHRzLnJlYyIsImFwYXNyLnRhc2siLCJhcHR0cy5ydHR0cyIsImFwYXNyLm9mZnNyIl0sImlzcyI6Imh0dHA6XC9cLzE5Mi4xNjguMjAwLjE5OTo4MTg0IiwiZXhwIjoxNzIyNDA0NjQzLCJpYXQiOjE3MjIzMTgyNDMsInRlbmFudCI6IjAwMDAwMDAwMDAwMCJ9.NR07Qjdi_aWevZF_XIaH-8Ej3QYYM2EKnm9zUkcDWj9IJzNgVzQCn0iSFSndyi_6U3DbbeMlnO3Dq2wpcvJLgoDE2CxuT663GHyrWHuT4LZ5IgqwLrLVPLjt9zMUl1_mELqTUh-JTbi5aMomlB-vFpBQB7FAnIMSuKZzHXE8ouB5vKZv-lLVQvV1KHquSIF6h9cw0g8GFYSBrScdCcfMQRQnxGYIyNe7qO4x_qy9d0pJdHhCIZ6goRaR3e3oUmkKYuT7y6riUirNp0QKwrHGkyfFLNZ1LBtxRG59beVgj8u0kJf60y5alrxPOSgQ0PujOizo66fchIGhdtepubVX4g"},
	})
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	//关闭连接
	defer conn.Close()
	go receiveHandler(conn)

	paramCh := make(chan interface{})
	// 发送请求参数
	go func() {
		paramCh <- 1

		params := url2.Values{}
		params.Add("appId", "85c62744-a4a1-4e08-b865-1796cc2ec622")
		params.Add("speechId", "BBBB")
		params.Add("modelName", "default")
		params.Add("format", "audio/16LE;rate=8000;no-header")
		params.Add("isAddPunct", "true")
		params.Add("isTransDigit", "true")
		params.Add("EmotionDetect", "false")
		params.Add("outputUndecided", "true")
		params.Add("outputFormat", "true")

		log.Println(params.Encode())

		err := conn.WriteMessage(websocket.BinaryMessage, []byte(params.Encode()))
		if err != nil {
			log.Fatal(err)
		}
	}()

	for {
		select {
		// 已经发送请求参数后，发送录音文件
		case <-paramCh:
			log.Println("开始发送语音流数据")
			file, err := os.Open("a_l.wav")
			if err != nil {
				log.Fatal(err.Error())
			}
			defer file.Close()
			reader := bufio.NewReaderSize(file, 4096)

			sum := 0
			for {
				time.Sleep(time.Millisecond * 8000 * 16 / 8 / 4096 * 100)
				// 定义每次读取发送4096字节
				buff := make([]byte, 4096)
				n, err := reader.Read(buff)
				sum += n
				log.Printf("已经读取了 %d 字节", sum)
				if err != nil {
					if err == io.EOF {
						log.Println("读取结束")

						conn.WriteMessage(websocket.TextMessage, []byte("{\"status\":\"end\"}"))
						break
					}
				}
				err = conn.WriteMessage(websocket.BinaryMessage, buff)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}
