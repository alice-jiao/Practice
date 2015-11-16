package chat

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/goconfig"
	"gopkg.in/redis.v3"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestRedisConn(t *testing.T) {
	testKey := "test_go_redis"
	testVal := "hello"
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := client.Set(testKey, testVal, 1*time.Minute).Err()
	if err != nil {
		t.Error(err)
	}
	value, err := client.Get(testKey).Result()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(testVal, "====", string(value))
	if testVal != string(value) {
		t.Error("value getting from redis is not equal to origin")
	}

	defer client.Close()
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

func TestEncode(t *testing.T) {
	var oriMsg chatMsg
	oriMsg.User = "localhost:1234"
	oriMsg.Time = time.Now().Unix()
	oriMsg.Msg = "hello world"
	fmt.Println("before encode:", oriMsg)

	msg, err := oriMsg.jsonEncode()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("after encode", msg)

	cmsg := new(chatMsg)
	json.Unmarshal([]byte(msg), cmsg)
	fmt.Println("after decode:", cmsg)
}

func TestPubSub(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	fmt.Println("client start")
	// defer client.Close()

	pubMsg := &chatMsg{"localhost:1234", time.Now().Unix(), "hello, i am online !"}
	jsonMessage, err := pubMsg.jsonEncode()
	if err != nil {
		t.Error(err)
	}

	pubsub, err := client.Subscribe(channel)
	if err != nil {
		t.Error(err)
	}
	defer pubsub.Close()
	fmt.Println("sub success,", pubsub)

	go func(*redis.PubSub) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()

		subMsg := new(redis.Message)
		recvMsg := new(chatMsg)
		i := 0
		for {
			fmt.Println("runing ", i)
			msgi, err := pubsub.Receive()
			i += 1
			if err != nil {
				t.Error(err)
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
			fmt.Println("recv msg:", subMsg)

			json.Unmarshal([]byte(subMsg.Payload), recvMsg)
			fmt.Println(recvMsg.User, " ", time.Unix(recvMsg.Time, 0).Format(time.ANSIC), " :")
			fmt.Println(recvMsg.Msg)
		}
	}(pubsub)

	fmt.Println("ready to pub,", jsonMessage)
	err = client.Publish(channel, jsonMessage).Err()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("pub success")

	pubMsg.Time = time.Now().Unix()
	pubMsg.Msg = "can u see me ?"
	jsonMessage, err = pubMsg.jsonEncode()
	if err != nil {
		t.Error(err)
	}
	err = client.Publish(channel, jsonMessage).Err()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("pub success")

	time.Sleep(time.Second * 3)
	fmt.Println("test end")
}
