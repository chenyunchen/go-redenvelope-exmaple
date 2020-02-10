package service

import (
	"gopkg.in/redis.v3"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"
)

type RedisService struct {
	client *redis.Client
	sha string
}

func NewRedisService() *RedisService {
	options := redis.Options{
		Network:  "tcp4",
		Addr:     "127.0.0.1:16379",
		DB:       1,
		PoolSize: 10,
	}
	client := redis.NewClient(&options)

	return &RedisService{
		client: client,
	}
}

func (s *RedisService) LoadLuaScript(path string) (err error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	sha, err := s.client.ScriptLoad(string(fileBytes)).Result()
	if err != nil {
		return
	}
	s.sha = sha

	return
}

func (s *RedisService) Eval(keys, args []string) (res interface{}, err error) {
  return s.client.EvalSha(s.sha, keys, args).Result()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (s *RedisService) TT() (err error) {
	count := 1
	for u:=1; u<=100; u++ {
		for a := 1; a <= 1000 ; a++ {
			//_, err = s.client.LPush("u:"+string(u), "helloworld").Result()
			_, err = s.Eval([]string{"u:"+strconv.Itoa(u)}, []string{"push_message", "helloworld", strconv.Itoa(count)})
			count++
	 	}
	 }
	return nil
}