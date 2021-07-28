package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

type callbackPayload struct {
	IDs []int `json:"object_ids"`
}

var queue = Queue{}

func init() {
	log.SetLevel(log.DebugLevel)

	err := queue.connect()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Connection established!")
}

func main() {
	http.HandleFunc("/callback", handleCallback) // todo env variable
	log.Fatal(http.ListenAndServe(":9090", nil)) // todo env variable
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(extendError(err, "handleCallback: failed to read body"))
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
		}

		var callbackObject callbackPayload
		err = json.Unmarshal(reqBody, &callbackObject)
		if err != nil {
			log.Error(extendError(err, "handleCallback: failed to parse body"))
			http.Error(w, "Error to parse body", http.StatusInternalServerError)
		}

		log.Debug("Received: ", callbackObject.IDs)

		if len(callbackObject.IDs) > 0 {
			err = publishIDs(callbackObject.IDs)
			if err != nil {
				log.Error(extendError(err, "handleCallback: failed to send message to queue"))
				http.Error(w, "Error to register IDs", http.StatusInternalServerError)
			}

			log.Info("Successfully registered: ", callbackObject.IDs)
		}

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid request method for this endpoint", http.StatusMethodNotAllowed)
	}

}

func publishIDs(IDs []int) error {
	var wg sync.WaitGroup

	// these channels are used to handle concurrent errors
	// ref: https://golangcode.com/errors-in-waitgroups/
	errors := make(chan error)
	wgDone := make(chan bool)

	// for each id create an independent goroutine to save it on the queue
	for _, ID := range IDs {
		wg.Add(1)
		go func(_id int) {
			defer wg.Done()
			err := queue.publish(_id)
			if err != nil {
				errors <- err
			}
		}(ID)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-errors:
		close(errors)
		return err
	}

	return nil
}
