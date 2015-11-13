package chat

import (
	"fmt"
	"github.com/Unknwon/goconfig"
	"os"
	"redis"
	"strconv"
	"testing"
)

func TestRedisConn(t *testing.T) {
	testKey := "test_go_redis"
	testVal := "hello"
	client, err := redis.NewSynchClient()
	if err != nil {
		t.Error(err)
	}

	client.Set(testKey, []byte(testVal))
	value, err := client.Get(testKey)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(testVal, "====", string(value))
	if testVal != string(value) {
		t.Error("value getting from redis is not equal to origin")
	}
}

func TestInit(t *testing.T) {
	channel := "chat_channel"
	spec := redis.DefaultSpec().Host("127.0.0.1").Port(6379)
	client, err := redis.NewAsynchClientWithSpec(spec)
	if err != nil {
		t.Error(err)
	}

	subClient, err := redis.NewPubSubClientWithSpec(spec)
	if err != nil {
		t.Error(err)
	}

	subClient.Subscribe(channel)
	client.Publish(channel, []byte("hi, i am online."))
	fmt.Println("client init done")
}

func TestConfig(t *testing.T) {
	config, err := goconfig.LoadConfigFile("chat_conf.ini")
	if err != nil {
		t.Error(err)
	}

	host, err := config.GetValue("redis", "host")
	if err != nil {
		t.Error(err)
	}

	port, err := config.GetValue("redis", "port")
	if err != nil {
		t.Error(err)
	}

	fmt.Println("redis config, host:", host, " port:", port)

	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Error(err)
	}

	cfg := &redisConfig{host: host, port: portNum}
	fmt.Println("create config obj:", cfg)
	fmt.Println("redis config, host:", cfg.host, " port:", cfg.port)
}

func TestUsername(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("failed get hostname")
	}
	user = fmt.Sprintf("%s:%d", hostname, os.Getpid())
	fmt.Println("user:", user)
}
