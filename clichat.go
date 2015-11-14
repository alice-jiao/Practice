package chat

import (
	"bufio"
	"encoding/json"
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
	User string
	Time int64
	Msg  string
}

var channel = "chat_channel"
var user string

func main() {
	client, subClient := Init()
	defer client.Quit()
	defer subClient.Quit()
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

		rcvCnt, err := client.Publish(channel, []byte(jsonMessage))
		if err != nil {
			fmt.Printf("Error on Publish - %s", err)
		} else {
			fmt.Printf("Message sent to %d subscribers\n", rcvCnt)
		}
	}
	fmt.Println(subClient)
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

	cfg := &redisConfig{host: host, port: port}
	return cfg, nil
}
