package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	gopxgrid "github.com/vkumov/go-pxgrid"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func dataPrinter(dataChan <-chan *gopxgrid.Message[gopxgrid.SessionTopicMessage]) {
	for data := range dataChan {
		if data.Err == nil && data.UnmarshalError == nil {
			bts, _ := json.Marshal(data.Body)
			logger.Info("Message", "body", string(bts))
		} else if data.UnmarshalError != nil {
			logger.Error("Failed to unmarshal message", "err", data.UnmarshalError)
		} else {
			logger.Error("Failed to read message", "err", data.Err)
		}
	}
}

func getX509Pair(certFile, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func loadCertPool(caFolder string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	files, err := os.ReadDir(caFolder)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		bts, err := os.ReadFile(filepath.Join(caFolder, file.Name()))
		if err != nil {
			return nil, err
		}
		if !pool.AppendCertsFromPEM(bts) {
			logger.Warn("Failed to append certificate", "file", file.Name())
		}
	}
	return pool, nil
}

func pxGridConfigFromFlags() *gopxgrid.PxGridConfig {
	c := gopxgrid.NewPxGridConfig()

	var (
		host string
		port int
		dns  string
	)
	flag.StringVar(&host, "host", "", "Host name (multiple accepted)")
	flag.IntVar(&port, "port", 8910, "Control port (optional)")

	// flag.StringVar(&c.filter, "f", "", "Server Side Filter (optional)")

	flag.StringVar(&c.NodeName, "n", "", "Node name")
	flag.StringVar(&c.Description, "d", "", "Description (optional)")

	certCfg := struct {
		certFile string
		keyFile  string
		password string
		caFolder string
	}{}

	flag.StringVar(&certCfg.certFile, "c", "", "Client certificate chain .pem filename (not required if password is specified)")
	flag.StringVar(&certCfg.keyFile, "k", "", "Client key unencrypted .key filename (not required if password is specified)")
	flag.StringVar(&certCfg.password, "w", "", "Password (not required if client certificate is specified)")
	flag.StringVar(&certCfg.caFolder, "s", "", "Folder with CA certificates (optional)")
	flag.StringVar(&dns, "dns", "", "DNS server (optional)")
	flag.BoolVar(&c.TLS.InsecureTLS, "insecure", false, "Insecure skip validation")
	flag.Parse()

	c.AddHost(host, port)
	c.Auth.Username = c.NodeName
	if dns != "" {
		c.SetDNS(dns, gopxgrid.DefaultINETFamilyStrategy)
	}

	if certCfg.certFile != "" && certCfg.keyFile != "" {
		cert, err := getX509Pair(certCfg.certFile, certCfg.keyFile)
		if err != nil {
			logger.Error("Failed to load client certificate", "err", err)
			os.Exit(1)
		}
		c.TLS.ClientCertificate = cert
	} else if certCfg.password != "" {
		c.Auth.Password = certCfg.password
	} else {
		logger.Error("Client certificate or password is required")
		os.Exit(1)
	}

	if certCfg.caFolder != "" {
		pool, err := loadCertPool(certCfg.caFolder)
		if err != nil {
			logger.Error("Failed to load CA pool", "err", err)
			os.Exit(1)
		}
		c.SetCA(pool)
		logger.Info("CA pool loaded", "len", len(pool.Subjects()))
	}

	return c
}

func main() {
	config := pxGridConfigFromFlags()
	config.SetLogger(logger)

	logger.Info("Connecting to pxGrid", slog.Any("config", config))

	control, err := gopxgrid.NewPxGridConsumer(config)
	if err != nil {
		logger.Error("Failed to create pxGrid consumer", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Info("Disconnecting from pxGrid")
		cancel()
		done <- true
	}()

	for {
		logger.Info("Activating account")
		res, err := control.Control().AccountActivate(ctx)
		if err != nil {
			logger.Error("Failed to activate account", "err", err)
			os.Exit(1)
		}
		if res.IsEnabled() {
			break
		}
		time.Sleep(30 * time.Second)
	}

	logger.Info("Account activated")
	sd := control.SessionDirectory()
	err = sd.CheckNodes(ctx)
	if err != nil {
		logger.Error("Failed to check nodes", "err", err)
		os.Exit(1)
	}

	logger.Info("Subscribing to session topic")
	ps, err := sd.Properties().WSPubsubService()
	if err != nil {
		logger.Error("Failed to get PubSub service name", "err", err)
		os.Exit(1)
	}

	err = control.PubSub(ps).UpdateSecrets(ctx)
	if err != nil {
		logger.Error("Failed to update secrets", "err", err)
		os.Exit(1)
	}

	sub, err := sd.OnSessionTopic().Subscribe(ctx)
	if err != nil {
		logger.Error("Failed to subscribe to session topic", "err", err)
		os.Exit(1)
	}

	go dataPrinter(sub.C)

	// Setup abort channel
	logger.Info("Press <Ctrl-c> to disconnect...")
	<-done

	logger.Info("Unsubscribing from session topic")
	err = sub.Unsubscribe()
	if err != nil {
		logger.Error("Failed to unsubscribe", "err", err)
		os.Exit(1)
	}
	logger.Info("Unsubscribed, exiting...")
}
