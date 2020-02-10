package service

type Service struct {
	Redis *RedisService
	RedEnvelope *RedEnvelopeService
}

func New() *Service {
	redisService := NewRedisService()
	redisService.LoadLuaScript("./main.lua")

	return &Service{
		Redis: redisService,
		RedEnvelope: NewRedEnvelopeService(redisService),
	}
}