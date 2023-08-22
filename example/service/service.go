package service

import (
	m "github.com/sandrolain/gomsvc/example/models"
	"github.com/sandrolain/gomsvc/pkg/body"
	c "github.com/sandrolain/gomsvc/pkg/client"
)

type HB = m.HelloBody

var SayHello = c.CreateClient[HB, m.HelloData](c.ClientOptions[HB]{
	ContentType: body.TypeMsgpack,
	Method:      "POST",
	Url:         "http://localhost:3000/hello",
})
