package rmqpublisher

import (
	"crypto/tls"
	"errors"
	"fmt"

	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
	//"github.com/davecgh/go-spew/spew"
)

type Publisher struct {
	Chan       *amqp.Channel
	Conn       *amqp.Connection
	QueueName  string
	Exchange   string
	RoutingKey string
	URL        string
	Tag        string
	InMsg      chan amqp.Delivery
	Quit       chan bool
	Done       chan error
}

type PublisherHandler func(msgq <-chan amqp.Delivery, quit <-chan bool, done chan<- error, p *Publisher)

func NewPublisher(amqpURI, exchange, exchangeType, queueName, key, ctag string, ph PublisherHandler, sslcfg *tls.Config) (*Publisher, error) {

	// make a new publisher
	p := Publisher{}
	var err error
	// Create a channel on which the publisher receives data
	p.InMsg = make(chan amqp.Delivery)
	p.Quit = make(chan bool)
	p.Done = make(chan error)

	if sslcfg != nil {
		p.Conn, err = amqp.DialTLS(amqpURI, sslcfg)
	} else {
		//Trace.Printf("dialing %q", amqpURI)
		p.Conn, err = amqp.Dial(amqpURI)
	}

	if err != nil {
		msg := fmt.Sprintf("Error connecting to rabbit server (%s): %s", amqpURI, err)
		log.Error(msg)
		err := errors.New(msg)
		return nil, err
	} else {
		ch, err := p.Conn.Channel()
		if err != nil {
			msg := fmt.Sprintf("Error creating channel for rabbit server (%s): %s", amqpURI, err)
			log.Error(msg)
			err = errors.New(msg)
			return nil, err
		} else {

			p.Chan = ch
			p.QueueName = queueName
			p.Exchange = exchange
			p.RoutingKey = key
			p.URL = amqpURI
			p.Tag = "rabbitrelay"
		}
	}
	go ph(p.InMsg, p.Quit, p.Done, &p)

	return &p, nil
}

func (p *Publisher) Shutdown() error {

	log.Trace("Start")

	if err := p.Chan.Cancel(p.Tag, true); err != nil {
		log.Error("Publisher cancel failed: %s", err)
		return fmt.Errorf("Publisher cancel failed: %s", err)
	}

	if err := p.Conn.Close(); err != nil {
		log.Error("AMQP connection close error: %s", err)
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	// Tell the handler to quit
	p.Quit <- true

	// Wait on the handler to quit (blocks)
	<-p.Done

	log.Trace("End")
	return nil
}

func (p *Publisher) PublishMessage(msg amqp.Delivery) {

	//log.Trace("Started")
	err := p.Chan.Publish(p.Exchange, msg.RoutingKey, false, false,
		amqp.Publishing{
			Headers:         msg.Headers,
			ContentType:     msg.ContentType,
			ContentEncoding: msg.ContentEncoding,
			Body:            msg.Body,
			DeliveryMode:    msg.DeliveryMode,
			Priority:        msg.Priority,
			Type:            msg.Type,
			AppId:           msg.AppId,
		},
	)
	//log.Trace("Message Sent")

	if err != nil {
		log.Error("Error publishing to RabbitMQ server (", p.URL, "): ", err)
	}
}
