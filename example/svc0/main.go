package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sandrolain/gomsvc/example/models"
	h "github.com/sandrolain/gomsvc/pkg/api"
	"github.com/sandrolain/gomsvc/pkg/api/client"
	"github.com/sandrolain/gomsvc/pkg/redislib"
	"github.com/sandrolain/gomsvc/pkg/repo"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Config struct {
	Redis redislib.EnvClientConfig
}

func main() {
	svc.Service(svc.ServiceOptions{
		Name:    "example",
		Version: "1.2.3",
	}, func(cfg Config) {
		fmt.Printf("cfg: %v\n", cfg)

		go redis(cfg)
		// go server(cfg)

		// time.Sleep(time.Second)
		// go httpClient()
	})
}

type Data struct {
	Firstname string
	Lastname  string
}

func redis(cfg Config) {
	svc.PanicIfError(
		redislib.Connect(redislib.ClientOptionsFromEnvConfig(cfg.Redis)),
	)

	// pub := redislib.Publisher[Data]("signup", redislib.PublisherConfig{Type: "signup"})
	pub := svc.PanicWithError(
		redislib.NewStreamPublisher[Data](redislib.StreamPublisherConfig{
			Stream: "mystream",
			Type:   "signup",
		}),
	)

	t := time.NewTicker(time.Millisecond * 100)
	for {
		<-t.C
		err := pub.Publish(Data{
			Firstname: "John",
			Lastname:  "Doe",
		})
		fmt.Printf("err: %v\n", err)
	}
}

func server(cfg Config) {

	h.SetLogger(svc.Logger())

	h.Authorize(func(ctx *fiber.Ctx) error {
		fmt.Printf("ctx: %v\n", ctx)
		if _, ok := ctx.GetReqHeaders()["X-Token"]; !ok {
			return fmt.Errorf("Un-auth")
		}
		return nil
	})

	// h.FilterError(func(re *h.RouteError) *h.RouteError {
	// 	fmt.Printf("re: %v\n", re)
	// 	if re.Status == 400 {
	// 		errors := re.Error.(validator.ValidationErrors)
	// 		fmt.Printf("errors: %v\n", errors)
	// 		re.Body = []byte(fmt.Sprintf("Bad Request!!!! %v", re.Ctx.Request().URI()))
	// 	}
	// 	return re
	// })

	// h.Handle("POST", "/hello", api.DataHandler(func(d *models.HelloData, c *fiber.Ctx) error {
	// 	fmt.Printf("d: %+v\n", d)
	// 	c.JSON(d)
	// 	return nil
	// }))

	// h.ListenPort()
}

type Car struct {
	Id    primitive.ObjectID `bson:"_id,omitempty"`
	Brand string             `bson:"brand"`
	Model string             `bson:"model"`
	Year  int                `bson:"year"`
}

func db() {
	conn, err := repo.Connect("mongodb://root:mypassword@localhost:27017", "msvc")
	if err != nil {
		panic(err)
	}

	cars := repo.NewRepo[Car, primitive.ObjectID](conn, "cars")
	cars.SetIdGenerator(func() (primitive.ObjectID, error) {
		return primitive.NewObjectID(), nil
	})

	car := Car{
		Brand: "Fiat",
		Model: "500",
		Year:  1990,
	}
	// err = cars.ApplyId(&car)
	// fmt.Printf("err: %v\n", err)

	id, err := cars.Save(car)
	fmt.Printf("id: %v\n", id)
	fmt.Printf("err: %v\n", err)

	count, err := cars.Delete(car)
	fmt.Printf("count: %v\n", count)
	fmt.Printf("err: %v\n", err)

	id, _ = primitive.ObjectIDFromHex("64a875b0ac2e928768307bd2")
	res, err := cars.Get(id)
	fmt.Printf("res: %v\n", res)

	resl, err := cars.Find(map[string]interface{}{})
	fmt.Printf("resl: %+v\n", resl)
	fmt.Printf("err: %v\n", err)
}

func httpClient() {
	res, err := client.PostJSON[models.HelloBody](context.Background(), "http://localhost:3000/hello", client.Init{
		Headers: map[string]string{
			"X-Token": "my-token",
		},
		Query: map[string]string{
			"num": "123",
		},
		Body: models.HelloBody{
			Type:   "Hello",
			Salary: 1234,
		},
	})
	fmt.Printf("res: %v\n", res)
	fmt.Printf("err: %v\n", err)
}
