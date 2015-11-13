package chat

import (
	"bufio"
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
	time time.Time
	msg  []byte
}

var channel = "chat_channel"
var user string

func main() {
	client, subClient := Init()
	welcome()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		if input == "\n" {
			os.Exit(0)
		}

		rcvCnt, err := client.Publish(channel, []byte(input))
		if err != nil {
			fmt.Printf("Error on Publish - %s", err)
		} else {
			fmt.Printf("Message sent to %d subscribers\n", rcvCnt)
		}
	}
	fmt.Println(subClient)
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

	// cfg := new(redisConfig)
	// cfg.host = host
	// cfg.port = port
	cfg := &redisConfig{host: host, port: port}
	return cfg, nil
}
