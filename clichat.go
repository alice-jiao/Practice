package chat

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/Unknwon/goconfig"
	"os"
	"redis"
	"strconv"
	"time"
)

type redisConfig struct {
	host string
	port int
}

type chatMsg struct {
	user string
	time int64
	msg  string
}

var channel = "chat_channel"
var user string
var network bytes.Buffer
var enc = gob.NewEncoder(&network)
var dec = gob.NewDecoder(&network)

func (m *chatMsg) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	err := encoder.Encode(m.user)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(m.time)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(m.msg)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (m *chatMsg) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	err := decoder.Decode(&m.user)
	if err != nil {
		return err
	}
	err = decoder.Decode(&m.time)
	if err != nil {
		return err
	}
	return decoder.Decode(&m.msg)
}

func main() {
	client, subClient := Init()
	defer client.Quit()
	defer subClient.Quit()
	welcome()
	inputReader := bufio.NewReader(os.Stdin)
	var cmsg chatMsg
	cmsg.user = user
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		if input == "\n" {
			os.Exit(0)
		}

		cmsg.time = time.Now().Unix()
		cmsg.msg = input

		// rcvCnt, err := client.Publish(channel, []byte(cmsg))
		// if err != nil {
		// 	fmt.Printf("Error on Publish - %s", err)
		// } else {
		// 	fmt.Printf("Message sent to %d subscribers\n", rcvCnt)
		// }
	}
	fmt.Println(subClient)
}

// func Msg2Byte(cmsg chatMsg) []byte {

// }

func welcome() {
	fmt.Println("======================================")
	fmt.Println("=                                    =")
	fmt.Println("=      welcome to the chat room      =")
	fmt.Println("=                                    =")
	fmt.Println("======================================")
}

func Init() (redis.AsyncClient, redis.PubSubClient) {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("failed get hostname")
	}
	user = fmt.Sprintf("%s:%d", hostname, os.Getpid())

	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}

	spec := redis.DefaultSpec().Host(cfg.host).Port(cfg.port)

	client, err := redis.NewAsynchClientWithSpec(spec)
	if err != nil {
		panic(err)
	}

	subClient, err := redis.NewPubSubClientWithSpec(spec)
	if err != nil {
		panic(err)
	}

	subClient.Subscribe(channel)
	client.Publish(channel, []byte("hi, i am online."))
	return client, subClient
}

func getConfig() (*redisConfig, error) {
	config, err := goconfig.LoadConfigFile("chat_conf.ini")
	if err != nil {
		return nil, err
	}

	host, err := config.GetValue("redis", "host")
	if err != nil {
		return nil, err
	}

	portNum, err := config.GetValue("redis", "port")
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portNum)
	if err != nil {
		return nil, err
	}

	// cfg := new(redisConfig)
	// cfg.host = host
	// cfg.port = port
	cfg := &redisConfig{host: host, port: port}
	return cfg, nil
}
