package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type objectPayload struct {
	ID     int  `json:"id"`
	Online bool `json:"online"`
}

var queue = Queue{}
var client *http.Client

func init() {
	log.SetLevel(log.DebugLevel)

	err := queue.connect()
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Connection established!")

	client = &http.Client{
		// extended timeout since the API can take up to 4 seconds to reply
		Timeout: 5 * time.Second,
	}
}

func main() {
	maxConsumersArg := os.Getenv("MAX_CONSUMERS")
	maxConsumers, err := strconv.Atoi(maxConsumersArg)
	if err != nil {
		log.Fatal(extendError(err, "invalid MAX_CONSUMERS env variable"))
	}

	// pool of workers that will consume the queue concurrently
	consumers := make(chan struct{}, maxConsumers)

	for {
		consumers <- struct{}{}
		go func() { consume(consumers) }()
	}
}

func consume(kill <-chan struct{}) {
	messages, channel, err := queue.consume()
	if err != nil {
		log.Error(err)
		<-kill // signal to recreate another consumer after this crashes
		return
	}

	end := make(chan bool)

	go func() {
		for d := range messages {
			var msg Message
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Error(extendError(err, "failed to parse body"))
				err = channel.Nack(d.DeliveryTag, false, true)
				if err != nil {
					log.Error(extendError(err, "failed to nack"))
				}
				continue
			}

			log.Debugf("Received message: %v", msg)

			isOnline, timestamp, err := fetchObjectStatus(client, msg.ID)
			if err != nil {
				log.Error(extendError(err, "failed to obtain object status"))
				err = channel.Nack(d.DeliveryTag, false, true)
				if err != nil {
					log.Error(extendError(err, "failed to nack"))
				}
				continue
			}

			if isOnline {
				log.Debugf("ID %d is online", msg.ID)
			} else {
				log.Debugf("ID %d is NOT online", msg.ID)
			}

			if isOnline {
				err = InsertObject(msg.ID, timestamp)
				if err != nil {
					log.Error(err)
					err = channel.Nack(d.DeliveryTag, false, true)
					if err != nil {
						log.Error(extendError(err, "failed to nack"))
					}
					continue
				}
				log.Debugf("Object ID=%d inserted on db", msg.ID)
			}

			// acknowledge the message so it can be removed from the queue
			err = channel.Ack(d.DeliveryTag, false)
			if err != nil {
				log.Error(extendError(err, "failed to ack a delivery"))
				<-kill
				return
			}
		}
	}()

	<-end
}

// Fetch object status from `/objects/:id` endpoint (online: status = true)
func fetchObjectStatus(client *http.Client, ID int) (bool, time.Time, error) {
	var timestamp time.Time

	resp, err := client.Get(fmt.Sprintf("%s/objects/%d", os.Getenv("TESTER_ENDPOINT"), ID))
	if err != nil {
		return false, timestamp, extendError(err, "failed to fetch status endpoint")
	}

	// last seen timestamp obtained after the object status endpoint returned
	timestamp = time.Now()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, timestamp, extendError(err, "failed to read body")
	}

	var object objectPayload
	err = json.Unmarshal(body, &object)
	if err != nil {
		return false, timestamp, extendError(err, "failed to parse body")
	}

	err = resp.Body.Close()
	if err != nil {
		return false, timestamp, extendError(err, "failed to close body")
	}

	return object.Online, timestamp, nil
}
