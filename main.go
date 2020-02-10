package main

import (
	"encoding/json"
	"fmt"
	"go-redis-lua/service"
	"os"
	"strconv"
	"time"
)

type Message struct {
	MessageID string `json:"msgID"`
	Content   string `json:"content"`
}

func main() {
	if len(os.Args) < 2 {
		return
	}

	s := service.New()

	switch os.Args[1] {
	case "get_redenvelope":
		userID := strconv.FormatUint(uint64(time.Now().Unix()), 10)
		res, err := s.RedEnvelope.GetRedEnvelope(userID)
		if err != nil {
			fmt.Printf("GetRedEnvelope|Fail|%v\n", err)
		}
		fmt.Printf("UserID: %v, Money: %v", res.UserID, res.Money)
	case "set_redenvelope":
		err := s.RedEnvelope.SetRedEnvelopes(1000, 10)
		if err != nil {
			fmt.Printf("SetRedEnvelope|Fail|%v\n", err)
		}
	case "push_message":
		userID := "1558578089"
		msgID := strconv.FormatUint(uint64(time.Now().Unix()), 10)
		message := Message{
			MessageID: msgID,
			Content:   "hello world",
		}
		b, err := json.Marshal(message)
		if err != nil {
			return
		}

		_, err = s.Redis.Eval([]string{"u:" + userID}, []string{"push_message", string(b), msgID})
		if err != nil {
			fmt.Printf("PushMessage|Fail|%v\n", err)
		}
	case "get_message":
		userID := "1558578089"
		res, err := s.Redis.Eval([]string{"u:" + userID}, []string{"get_message"})
		if err != nil {
			fmt.Printf("GetMessage|Fail|%v\n", err)
		}
		for _, strMsg := range res.([]interface{}) {
			fmt.Println(string([]byte(strMsg.(string))))
		}
	case "set_finish":
		userID := "1558578089"
		head, err := s.Redis.Eval([]string{"u:" + userID}, []string{"set_finish", os.Args[2]})
		if err != nil {
			fmt.Printf("SetFinish|Fail|%v\n", err)
		}
		fmt.Printf("Head: %v", head)
	}
}
