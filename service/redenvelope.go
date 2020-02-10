package service

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"math/rand"
	"strconv"
	"time"
)

type RedEnvelope struct {
	UserID uint32 `json:"userID"`
	Money  string `json:"money"`
}

type RedEnvelopeService struct {
	redisService *RedisService
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func NewRedEnvelopeService(redisService *RedisService) *RedEnvelopeService {
	return &RedEnvelopeService{
		redisService: redisService,
	}
}

func (s *RedEnvelopeService) GetRedEnvelope(userID string) (re RedEnvelope, err error) {
	rID := strconv.FormatUint(uint64(time.Now().Unix()), 10)
	res, err := s.redisService.client.EvalSha(s.redisService.sha, []string{"r:" + rID}, []string{"get_redenvelope", userID}).Result()
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(res.(string)), &re)
	if err != nil {
		return
	}

	return
}

func (s *RedEnvelopeService) SetRedEnvelopes(money, count int) (err error) {
	redenvelopes := make([]string, count)
	rates := []float64{0.6, 0.7, 0.8, 0.9, 0.99}

	moneyDec := decimal.NewFromFloat(float64(money))
	countDec := decimal.NewFromFloat(float64(count))
	rateDec := decimal.NewFromFloat(float64(rates[rand.Intn(5)]))

	average := moneyDec.Div(countDec).Truncate(2)
	gap := average.Div(countDec.Sub(decimal.NewFromFloat(1))).Truncate(2)

	tmp := average.Mul(rateDec).Truncate(2)
	min := average.Sub(tmp)
	max := average.Add(tmp)

	for i, _ := range redenvelopes {
		if i == count-i-1 {
			redenvelopes[i] = average.String()
			break
		}
		if i > count-i-1 {
			break
		}
		tmp = gap.Mul(decimal.NewFromFloat(float64(i)))
		redenvelopes[i] = min.Add(tmp).String()
		redenvelopes[count-i-1] = max.Sub(tmp).String()

	}
	r0, err := decimal.NewFromString(redenvelopes[0])
	if err != nil {
		return
	}
	redenvelopes[0] = r0.Add(moneyDec.Sub(average.Mul(countDec))).String()

	for i, money := range redenvelopes {
		rIndex := rand.Intn(count)
		redenvelopes[i] = redenvelopes[rIndex]
		redenvelopes[rIndex] = money
	}

	for _, money := range redenvelopes {
		r := RedEnvelope{
			Money: money,
		}
		b, e := json.Marshal(r)
		if e != nil {
			err = e
			return
		}

		_, err = s.redisService.client.LPush("redenvelope", string(b)).Result()
		if e != nil {
			err = e
			return
		}
	}
	return
}
