package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
)

var errorCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "mesos",
	Subsystem: "collector",
	Name:      "errors_total",
	Help:      "Total number of internal mesos-collector errors.",
})

func init() {
	// Only log the warning severity or above.
	log.SetLevel(log.ErrorLevel)

	prometheus.MustRegister(errorCounter)
}

func getX509CertPool(pemFiles []string) *x509.CertPool {
	pool := x509.NewCertPool()
	for _, f := range pemFiles {
		content, err := ioutil.ReadFile(f)
		if err != nil {
			log.WithField("error", err).Fatal("x509 certificate pool error")
		}
		ok := pool.AppendCertsFromPEM(content)
		if !ok {
			log.WithField("file", f).Fatal("Error parsing .pem file")
		}
	}
	return pool
}

func mkHTTPClient(url string, timeout time.Duration, auth authInfo, certPool *x509.CertPool) *httpClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: auth.skipSSLVerify},
	}

	// HTTP Redirects are authenticated by Go (>=1.8), when redirecting to an identical domain or a subdomain.
	// -> Hijack redirect authentication, since hostnames rarely follow this logic.
	var redirectFunc func(req *http.Request, via []*http.Request) error
	if auth.username != "" && auth.password != "" {
		// Auth information is only available in the current context -> use lambda function
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			req.SetBasicAuth(auth.username, auth.password)
			return nil
		}
	}

	client := &httpClient{
		http.Client{Timeout: timeout, Transport: transport, CheckRedirect: redirectFunc},
		url,
		auth,
		"",
	}

	if auth.strictMode {
		client.auth.signingKey = parsePrivateKey(client)
	}

	if version.Revision != "" {
		client.userAgent = fmt.Sprintf("mesos_exporter/%s (%s)", version.Version, version.Revision)
	} else {
		client.userAgent = fmt.Sprintf("mesos_exporter/%s", version.Version)
	}

	return client
}

func parsePrivateKey(httpClient *httpClient) []byte {
	if _, err := os.Stat(httpClient.auth.privateKey); os.IsNotExist(err) {
		buffer := bytes.NewBuffer([]byte(httpClient.auth.privateKey))
		var key mesosSecret
		if err := json.NewDecoder(buffer).Decode(&key); err != nil {
			log.WithFields(log.Fields{
				"key":   key,
				"error": err,
			}).Error("Error decoding prviate key")
			errorCounter.Inc()
			return []byte{}
		}
		httpClient.auth.username = key.UID
		httpClient.auth.loginURL = key.LoginEndpoint
		return []byte(key.PrivateKey)
	}
	absPath, _ := filepath.Abs(httpClient.auth.privateKey)
	key, err := ioutil.ReadFile(absPath)
	if err != nil {
		log.WithFields(log.Fields{
			"absPath": absPath,
			"error":   err,
		}).Error("Error reading prviate key")
		errorCounter.Inc()
		return []byte{}
	}
	return key
}

func csvInputToList(input string) []string {
	var entryList []string
	if input == "" {
		return entryList
	}
	sanitizedString := strings.Replace(input, " ", "", -1)
	entryList = strings.Split(sanitizedString, ",")
	return entryList
}

func main() {
	fs := flag.NewFlagSet("mesos-exporter", flag.ExitOnError)
	addr := fs.String("addr", ":9105", "Address to listen on")
	masterURL := fs.String("master", "", "Expose metrics from master running on this URL")
	masterState := fs.Bool("masterState", false, "Expose metrics derived from master state (polls /state, which can be expensive)")
	slaveURL := fs.String("slave", "", "Expose metrics from slave running on this URL")
	timeout := fs.Duration("timeout", 10*time.Second, "Master polling timeout")
	exportedTaskLabels := fs.String("exportedTaskLabels", "", "Comma-separated list of task labels to include in the corresponding metric")
	exportedSlaveAttributes := fs.String("exportedSlaveAttributes", "", "Comma-separated list of slave attributes to include in the corresponding metric")
	trustedCerts := fs.String("trustedCerts", "", "Comma-separated list of certificates (.pem files) trusted for requests to Mesos endpoints")
	strictMode := fs.Bool("strictMode", false, "Use strict mode authentication")
	username := fs.String("username", "", "Username for authentication")
	password := fs.String("password", "", "Password for authentication")
	loginURL := fs.String("loginURL", "https://leader.mesos/acs/api/v1/auth/login", "URL for strict mode authentication")
	logLevel := fs.String("logLevel", "error", "Log level")
	privateKey := fs.String("privateKey", "", "File path to certificate for strict mode authentication")
	skipSSLVerify := fs.Bool("skipSSLVerify", false, "Skip SSL certificate verification")
	vers := fs.Bool("version", false, "Show version")

	fs.Parse(os.Args[1:])

	if *vers {
		fmt.Println(version.Print("mesos_exporter"))
		os.Exit(0)
	}

	if *masterURL != "" && *slaveURL != "" {
		log.Fatal("Only -master or -slave can be given at a time")
	}

	// Getting logging setup with the appropriate log level
	logrusLogLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.WithField("logLevel", *logLevel).Fatal("invalid logging level")
	}
	if logrusLogLevel != log.ErrorLevel {
		log.SetLevel(logrusLogLevel)
		log.WithField("logLevel", *logLevel).Info("Changing log level")
	}

	log.Infoln("Starting mesos_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(version.NewCollector("mesos_exporter"))

	auth := authInfo{
		strictMode:    *strictMode,
		skipSSLVerify: *skipSSLVerify,
		loginURL:      *loginURL,
	}

	if *strictMode && *privateKey != "" {
		auth.privateKey = *privateKey
	} else {
		auth.privateKey = os.Getenv("MESOS_EXPORTER_PRIVATE_KEY")
		log.WithField("privateKey", auth.privateKey).Debug("strict mode, no private key, pulling from the environment")
	}

	if *username != "" {
		auth.username = *username
	} else {
		auth.username = os.Getenv("MESOS_EXPORTER_USERNAME")
		log.WithField("username", auth.username).Debug("auth with no username, pulling from the environment")
	}

	if *password != "" {
		auth.password = *password
	} else {
		auth.password = os.Getenv("MESOS_EXPORTER_PASSWORD")
		// NOTE it's already in the environment, so can be easily read anyway
		log.WithField("password", auth.password).Debug("auth with no password, pulling from the environment")
	}

	var certPool *x509.CertPool
	if *trustedCerts != "" {
		certPool = getX509CertPool(csvInputToList(*trustedCerts))
	}

	slaveAttributeLabels := csvInputToList(*exportedSlaveAttributes)
	slaveTaskLabels := csvInputToList(*exportedTaskLabels)

	switch {
	case *masterURL != "":
		collectors := []func(*httpClient) prometheus.Collector{
			newMasterCollector,
		}
		if *masterState {
			collectors = append(collectors, func(c *httpClient) prometheus.Collector {
				return newMasterStateCollector(c, slaveAttributeLabels)
			})
		}
		for _, f := range collectors {
			c := f(mkHTTPClient(*masterURL, *timeout, auth, certPool))
			if err := prometheus.Register(c); err != nil {
				log.WithField("error", err).Fatal("Prometheus Register() error")
			}
		}
		log.WithField("address", *addr).Info("Exposing master metrics")

	case *slaveURL != "":
		slaveCollectors := []func(*httpClient) prometheus.Collector{
			func(c *httpClient) prometheus.Collector {
				return newSlaveCollector(c)
			},
			func(c *httpClient) prometheus.Collector {
				return newSlaveMonitorCollector(c)
			},
			func(c *httpClient) prometheus.Collector {
				return newSlaveStateCollector(c, slaveTaskLabels, slaveAttributeLabels)
			},
		}

		for _, f := range slaveCollectors {
			c := f(mkHTTPClient(*slaveURL, *timeout, auth, certPool))
			if err := prometheus.Register(c); err != nil {
				log.WithField("error", err).Fatal("Prometheus Register() error")
			}
		}
		log.WithField("address", *addr).Info("Exposing slave metrics")

	default:
		log.Fatal("Either -master or -slave is required")
	}

	log.Info("Listening and serving ...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>Mesos Exporter</title></head>
            <body>
            <h1>Mesos Exporter</h1>
            <p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.WithField("error", err).Fatal("listen and serve error")
	}
}
