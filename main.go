package main

import (
	"github.com/rs/zerolog/log"
	"github.com/soldatov-s/go-garage-example/internal/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal().Err(err).Msg("run application")
	}
}

// func main() {
// 	ctx := context.Background()
// 	rabbitMQConnDriver := rabbitmqcon.NewDriver("amqp://guest:guest@rabbitmq:5672")
// 	poolRabbitMQConn := ucpool.OpenPool(ctx, rabbitMQConnDriver)

// 	poolRabbitMQConn.SetConnMaxLifetime(15 * time.Second)

// 	_, err := poolRabbitMQConn.Conn(context.Background())
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	conn, err := poolRabbitMQConn.Conn(context.Background())
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	rabbitMQConnDriverConn, rabbitMQConnDriverConnReleaser, err := conn.GrabConn(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Get conn as *amqp.Connection
// 	rabbitMQConn, ok := rabbitMQConnDriverConn.GetCi().(*amqp.Connection)
// 	if !ok {
// 		log.Fatal("failed typecast to *amqp.Connection")
// 	}
// 	rabbitMQChanDriver := rabbitmqchan.NewDriver(rabbitMQConn)
// 	poolRabbitMQChan := ucpool.OpenPool(ctx, rabbitMQChanDriver)
// 	poolRabbitMQChan.SetConnMaxLifetime(15 * time.Second)

// 	channel, err := poolRabbitMQChan.Conn(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	rabbitMQChanDriverChan, rabbitMQChanDriverChanReleaser, err := channel.GrabConn(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	rabbitMQChan, ok := rabbitMQChanDriverChan.GetCi().(*amqp.Channel)
// 	if !ok {
// 		log.Fatal("failed typecast to *amqp.Channel")
// 	}

// 	errExchangeDeclare := rabbitMQChan.ExchangeDeclare("testout.testqqq.events.dev", "direct", true,
// 		false, false,
// 		false, nil)

// 	rabbitMQChanDriverChanReleaser(errExchangeDeclare)

// 	rabbitMQConnDriverConnReleaser(nil)

// 	time.Sleep(5 * time.Second)
// 	channel.Close()

// 	fmt.Println("closed chan")

// 	time.Sleep(15 * time.Second)

// 	fmt.Println("15 seconds")

// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

// 	<-c
// }
