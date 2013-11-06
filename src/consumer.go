// This example declares a durable Exchange, an ephemeral (auto-delete) Queue,
// binds the Queue to the Exchange with a binding key, and consumes every
// message published to that Exchange with that routing key.
//
package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
    "mylib"
	"encoding/json"
	"encoding/xml"
	"os"
	"runtime"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchange     = flag.String("exchange", "router", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue        = flag.String("queue", "orders.fullfilment", "Ephemeral AMQP queue name")
	bindingKey   = flag.String("key", "orders.fullfilment", "AMQP binding key")
	consumerTag  = flag.String("consumer-tag", "", "AMQP consumer tag (should not be blank)")
	lifetime     = flag.Duration("lifetime", 0, "lifetime of process before shutdown (0s=infinite)")
	cpu         =  flag.Int("cpu", 1, "number of cpus to use")
	concurrentConsumers = flag.Int("concurrent-consumers", 1, "concurrent consumers to run")
)

func init() {
	flag.Parse()
}

func main() {
    // Dynamically setting num-of-cpus at runtime
	// runtime.GOMAXPROCS(runtime.NumCPU())
	
	// Using arguments passed by command line
	runtime.GOMAXPROCS(*cpu)
	c, err := NewConsumer(*uri, *exchange, *exchangeType, *queue, *bindingKey, *consumerTag, *concurrentConsumers)
	if err != nil {
		log.Fatalf("%s", err)
	}

	if *lifetime > 0 {
		log.Printf("running for %s", *lifetime)
		time.Sleep(*lifetime)
	} else {
		log.Printf("running forever")
		select {}
	}

	log.Printf("shutting down")

	if err := c.Shutdown(); err != nil {
		log.Fatalf("error during shutdown: %s", err)
	}
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string, concurrentConsumers int) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error

	log.Printf("dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}
	
	// How many messages to recieve at a time - this allows for fair dispatching
    c.channel.Qos(100,0,false)
    
	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,       // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	// Spin up some go-routines to handle workload
	// Spin many only if IO intensive (just my humble opinion)
	for i := 0; i < concurrentConsumers; i++ {
		go handle(deliveries, c.done)
	}
	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		order := mylib.Order{}
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		
		json.Unmarshal([]byte(d.Body), &order)
		createOrder := mylib.CreateOrder{Xmlns: "http://dc.artistservices.com/ClientID/v1/", Order: order}
		content := mylib.Content(createOrder)
		soapBody := mylib.SoapBody{Content: content}
		v := &mylib.Soap12Envelope{Xsi: "http://www.w3.org/2001/XMLSchema-instance", Xsd: "http://www.w3.org/2001/XMLSchema", Soap12: "http://www.w3.org/2003/05/soap-envelope", SoapBody: soapBody}
		enc := xml.NewEncoder(os.Stdout)
		enc.Indent("  ", "    ")
		if err := enc.Encode(v); err != nil {
			fmt.Printf("error: %v\n", err)
		}
		// lets fake an API post
		time.Sleep(1000 * time.Millisecond)
		
		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
