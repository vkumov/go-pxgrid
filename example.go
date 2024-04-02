package gopxgrid

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func dataPrinter(dataChan <-chan *Message[SessionTopicMessage]) {
	for data := range dataChan {
		if data.Err == nil && data.UnmarshalError == nil {
			bts, _ := json.Marshal(data.Body)
			log.Println("Message=" + string(bts))
		} else if data.UnmarshalError != nil {
			log.Println(data.UnmarshalError)
		} else {
			log.Println(data.Err)
		}
	}
}

func main() {
	config := NewPxGridConfig()
	control, err := NewPxGridConsumer(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
		done <- true
	}()

	// AccountActivate
	for {
		res, err := control.Control().AccountActivate(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if res.AccountState == AccountStateEnabled {
			break
		}
		time.Sleep(30 * time.Second)
	}

	sd := control.SessionDirectory()

	err = sd.CheckNodes(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sub, err := sd.OnSessionTopic().Subscribe(ctx)
	if err != nil {
		log.Fatal(err)
	}

	go dataPrinter(sub.C)

	// Setup abort channel
	log.Println("Press <Ctrl-c> to disconnect...")
	<-done

	err = sub.Unsubscribe()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Unsubscribed, exiting...")
}
