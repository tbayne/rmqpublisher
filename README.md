# rmqpublisher
A Go based library for publishing messages to a specified RabbitMQ queue

## Build Requriements
 * A Go Development Environment (see: https://golang.org/doc/install)

## Prerequisites:

  * amqp - A go library for accessing amqp servers
   * http://godoc.org/github.com/streadway/amqp
 
 * seelog - a Go logging framework
  * https://github.com/cihub/seelog/wiki

### Installing Prerequisite Libraries:
```
  go get github.com/streadway/amqp
  go get github.com/cihub/seelog

```

## Installing 
Due to the fact that 'go get' doesn't really understand gitlab based repositories 
that don't, in fact, begin with "gitlab." as part of the address, we have to 
download these libraries and install them locally:

```
    git clone https://git.synapse-wireless.com/hcap/rmqpubisher.git
    cd rmqpublisher
    go install
    
```
 
## Usage
The package defines one data structure (Consumer), one function type (ConsumerHandler), and
three exported functions (NewConsumer, PublishMessage, Shutdown)

```
  import (
    rp "synapse-wireless.com/rmqpublisher"
    rc "synapse-wireless.com/rmqex"
  )
```

### Data Types

#### Publisher 

```
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
```

#### PublisherHandler

```
    // The function signature of our handler
    type PublisherHandler func(msgq <-chan amqp.Delivery, quit <-chan bool, done chan<- error, p *Publisher)
```

### Functions

#### NewPublisher
Takes connection and Queue/Exchange parameters, establishers a connection to 
the specified RabbitMQ server, starts the provided Handler as a go routine and
then returns an initialized Publisher structure.

 * Note: If the sslcfg is parameter is not nil, the publisher will attempt an SSL Connection

```  
    func NewPublisher(amqpURI, exchange, exchangeType, queueName, key, ctag string, ph PublisherHandler, sslcfg *tls.Config) (*Publisher, error)
```

#### PublishMessage
Invoked via the returned Publisher structure, sends the msg (defined as an array of bytes)
to the specified RabbitMQ exchange.

 * Note that while you can call this method directly, generally it should be called
 from the handler

```
    func (p *Publisher) PublishMessage(msg []byte)
```

#### Shutdown
Invoked via the returned Publisher struct, to shutdown the rabbit connection  and the handler.

```
    func (p *Publisher) Shutdown() error
```

### Sample Handler
This is the publisher handler from rabbitrelaygo
```
  // Handler for messages to be published
  func publisherHandler(msgq <-chan amqp.Delivery, quit <-chan bool, done chan<- error, p *rp.Publisher) {
    for {
      select {
        case msg := <-msgq:
          p.PublishMessage(msg)
        case <-quit:
          msg := fmt.Sprintf("Shutting down handler for publisher (%s)", p.URL)
          log.Trace(msg)
          done <- nil
          return
      }
    }
  }
```

## Logging
Logging is handled via the seelog framework and configured by ./seelog.xml

 * Set the 'minlevel' to your minimum logging level.  One of:
  * "trace"
  * "debug"
  * "info"
  * "warn"
  * "error"
  * "critical"


Sample Logging Configuration file

```

<seelog minlevel="trace" maxlevel="critical">
    <outputs>
        <rollingfile type="size" filename="./rabbitrelay.log" maxsize="1000000" maxrolls="50" />
    </outputs>
</seelog>

```
