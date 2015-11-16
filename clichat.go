package main

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
	Type int
}

const (
	MsgTypeBroadcast = 0
	MsgTypeContent   = 1
)

var channel = "chat_channel"
var user string

func main() {
	//init client
	client := Init()
	defer client.Quit()

	pubsub, err := client.Subscribe(channel)
	if err != nil {
		panic(err)
	}
	defer pubsub.Close()

	//echo welcome title
	welcome()
	//begin listening
	go recvMsg(pubsub)

	//send online broadcast
	sendOnlineMsg(client)

	//get user input for publish
	inputReader := bufio.NewReader(os.Stdin)

	var cmsg chatMsg
	cmsg.User = user
	cmsg.Type = MsgTypeContent
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		if input == "\n" {
			sendOfflineMsg(client)
			os.Exit(0)
		}

		cmsg.Time = time.Now().Unix()
		cmsg.Msg = input
		publishMsg(client, &cmsg)
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

func recvMsg(pubsub *redis.PubSub) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in recvMsg", r)
		}
	}()

	//init message from redis
	subMsg := new(redis.Message)
	//init point of chatMsg for parsing message
	recvMsg := new(chatMsg)

	for {
		msgi, err := pubsub.Receive()
		if err != nil {
			panic(err)
		}

		switch msg := msgi.(type) {
		case *redis.Subscription:
			// Ignore.
			continue
		case *redis.Pong:
			// Ignore.
			continue
		case *redis.Message:
			subMsg = msg
		case *redis.PMessage:
			subMsg = &redis.Message{
				Channel: msg.Channel,
				Pattern: msg.Pattern,
				Payload: msg.Payload,
			}
		default:
			fmt.Errorf("redis: unknown message: %T", msgi)
			continue
		}

		json.Unmarshal([]byte(subMsg.Payload), recvMsg)
		printMsg(recvMsg)
	}
}

func printMsg(chatMsg *chatMsg) {
	if chatMsg.User == user {
		return
	}

	if chatMsg.Type == MsgTypeBroadcast {
		printSimpleMsg(chatMsg)
	} else {
		time := time.Unix(chatMsg.Time, 0)
		fmt.Printf("%s %d-%d-%d %d:%d:%d :\n", chatMsg.User, time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second())
		fmt.Println(chatMsg.Msg)
	}
}

func printSimpleMsg(chatMsg *chatMsg) {
	fmt.Println(chatMsg.Msg)
}

func sendOfflineMsg(client *redis.Client) {
	onlineMsg := &chatMsg{user, time.Now().Unix(), user + " is offline...", MsgTypeBroadcast}
	publishMsg(client, onlineMsg)
}

func sendOnlineMsg(client *redis.Client) {
	offlineMsg := &chatMsg{user, time.Now().Unix(), user + " is online!", MsgTypeBroadcast}
	publishMsg(client, offlineMsg)
}

func publishMsg(client *redis.Client, msg *chatMsg) {
	jsonMessage, err := msg.jsonEncode()
	if err != nil {
		fmt.Println("json encoding error,", err)
	}

	err = client.Publish(channel, jsonMessage).Err()
	if err != nil {
		panic(err)
	}
}
