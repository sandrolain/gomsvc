package service

import (
	m "github.com/sandrolain/gomscv/example/models"
	"github.com/sandrolain/gomscv/pkg/body"
	c "github.com/sandrolain/gomscv/pkg/client"
)

type HB = m.HelloBody

var SayHello = c.CreateClient[HB, m.HelloData](c.ClientOptions[HB]{
	ContentType: body.TypeMsgpack,
	Method:      "POST",
	Url:         "http://localhost:3000/hello",
})
