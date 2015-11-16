package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Unknwon/goconfig"
	"gopkg.in/redis.v3"
	"os"
	"strconv"
	"time"
)

type redisConfig struct {
	host string
	port int
}

type chatMsg struct {
	User string
	Time int64
	Msg  string
}

var channel = "chat_channel"
var user string

func main() {
	client := Init()
	defer client.Quit()

	welcome()
	inputReader := bufio.NewReader(os.Stdin)
	var cmsg chatMsg
	cmsg.User = user
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		if input == "\n" {
			os.Exit(0)
		}

		cmsg.Time = time.Now().Unix()
		cmsg.Msg = input
		jsonMessage, err := cmsg.jsonEncode()
		if err != nil {
			fmt.Println("json encoding error,", err)
		}

		err = client.Publish(channel, jsonMessage).Err()
		if err != nil {
			panic(err)
		}
		pubsub, err := client.Subscribe(channel)
		if err != nil {
			panic(err)
		}

		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			panic(err)
		}
		recvMsg := new(chatMsg)
		json.Unmarshal([]byte(msg.Payload), recvMsg)
		fmt.Println(recvMsg.User, " ", time.Unix(recvMsg.Time, 0), " :")
		fmt.Println(recvMsg.Msg)
	}

}

func (m *chatMsg) jsonEncode() (string, error) {
	res, err := json.Marshal(m)
	if err != nil {
		fmt.Println("json encoding error,", err)
		return "", err
	}

	return string(res), nil
}

func welcome() {
	fmt.Println("======================================")
	fmt.Println("=                                    =")
	fmt.Println("=      welcome to the chat room      =")
	fmt.Println("=                                    =")
	fmt.Println("======================================")
}

func Init() *redis.Client {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("failed get hostname")
	}
	user = fmt.Sprintf("%s:%d", hostname, os.Getpid())

	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}
	addr := cfg.host + ":" + strconv.Itoa(cfg.port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
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

	cfg := &redisConfig{host: host, port: port}
	return cfg, nil
}
