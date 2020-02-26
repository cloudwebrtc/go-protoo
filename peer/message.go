package peer

import (
	"math/rand"
	"time"
)

type AcceptFunc func(data map[string]interface{})
type RejectFunc func(errorCode int, errorReason string)

/*
* Request
{
  request : true,
  id      : 12345678,
  method  : 'chatmessage',
  data    :
  {
    type  : 'text',
    value : 'Hi there!'
  }
}
*/
type Request struct {
	Request bool                   `json:"request"`
	Id      int                    `json:"id"`
	Method  string                 `json:"method"`
	Data    map[string]interface{} `json:"data"`
}

/*
* Success response
{
	response : true,
	id       : 12345678,
	ok       : true,
	data     :
	{
	  foo : 'lalala'
	}
}
*/
type Response struct {
	Response bool                   `json:"response"`
	Id       int                    `json:"id"`
	Ok       bool                   `json:"ok"`
	Data     map[string]interface{} `json:"data"`
}

/*
* Error response
{
  response    : true,
  id          : 12345678,
  ok          : false,
  errorCode   : 123,
  errorReason : 'Something failed'
}
*/
type ResponseError struct {
	Response    bool   `json:"response"`
	Id          int    `json:"id"`
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"errorCode"`
	ErrorReason string `json:"errorReason"`
}

/*
* Notification
{
  notification : true,
  method       : 'chatmessage',
  data         :
  {
    foo : 'bar'
  }
}
*/
type Notification struct {
	Notification bool                   `json:"notification"`
	Method       string                 `json:"method"`
	Data         map[string]interface{} `json:"data"`
}

func RandInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min
}

func GenerateRandomNumber() int {
	return RandInt(1000000, 9999999)
}
