package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"fmt"
	"encoding/json"
	"strings"
	"time"
)

func getConfig(rawConfig string) map[string]int {
	var ConfigInfo map[string]int
	err := json.Unmarshal([]byte(rawConfig), &ConfigInfo)
	if err != nil {
		log.Printf("Error: %s", err)
	}
	return ConfigInfo
}

func handleHTTP(w http.ResponseWriter, req *http.Request, backendURL string,
	delayConfig map[string]int, blockConfig map[string]int) {

	log.Printf("Incoming request: %s, headers: %s",
		req.URL.Path, getHeadersJSON(req.Header))
	doDelay(delayConfig, req.URL.Path)
	if doBlock(blockConfig, req, w) {
		return
	}
	makeResponse(req, backendURL, w)
}

func getHeadersJSON(header map[string][]string) string {
	headers, _ := json.Marshal(header)
	return string(headers)
}

func doDelay(delayConfig map[string]int, path string) {
	if len(delayConfig) != 0 {
		for k, v := range delayConfig {
			if strings.Contains(path, k) {
				log.Printf("Sleeping %d seconds ...", v)
				dur, _ := time.ParseDuration(fmt.Sprintf("%ds", v))
				time.Sleep(dur)
				break
			}
		}
	}
}

func doBlock(blockConfig map[string]int, req *http.Request, w http.ResponseWriter) bool {
	if len(blockConfig) != 0 {
		for k, v := range blockConfig {
			if strings.Contains(req.URL.Path, k) {
				log.Printf("Block status code: %d", v)
				writeResponse(w, v, req.Body)
				return true
			}
		}
	}
	return false
}

func makeResponse(req *http.Request, backendURL string, w http.ResponseWriter) {
	client := &http.Client{}
	newReq, err := http.NewRequest(req.Method, getURL(req, backendURL), req.Body)
	copyHeader(req.Header, newReq.Header)
	resp, err := client.Do(newReq)

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(resp.Header, w.Header())
	log.Printf("Response status: %s, headers: %s",
		resp.Status, getHeadersJSON(resp.Header))

	writeResponse(w, resp.StatusCode, resp.Body)
}

func getURL(req *http.Request, backendURL string) (string) {
	return fmt.Sprintf("http://%s%s", backendURL, req.URL.Path)
}

func copyHeader(src, dst http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, body io.Reader)  {
	w.WriteHeader(statusCode)
	io.Copy(w, body)
}

func getCmdArgs() (lp int, bu string, dc map[string]int, bc map[string]int) {
	var localPort int
	var backendURL string
	var delayConfig string
	var blockConfig string
	flag.IntVar(&localPort, "localPort", 5050, "Local port")
	flag.StringVar(&backendURL, "backendURL", "127.0.0.1:8080", "Backend URL")
	flag.StringVar(&delayConfig, "delay_config", "{}",
		"Serialized dict for delaying backend endpoints"+
			" where keys are patterns of endpoints, values are delays in seconds,"+
			"for instance: \"{\\\"profile\\\": 35}\"")
	flag.StringVar(&blockConfig, "block_config", "{}",
		"Serialized dict for blocking backend endpoints"+
			" where keys are patterns of endpoints, values are response status codes,"+
			"for instance: \"{\\\"profile\\\": 404}\"")
	flag.Parse()
	log.Printf("Local port: %d, backend URL: \"%s\", delay config: \"%s\", block config: \"%s\"",
		localPort, backendURL, delayConfig, blockConfig)
	return localPort, backendURL, getConfig(delayConfig), getConfig(blockConfig)
}

func main() {
	localPort, backendURL, delayConfig, blockConfig := getCmdArgs()
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", localPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handleHTTP(w, req, backendURL, delayConfig, blockConfig)
		}),
	}
	log.Printf("Starting REST proxy on local port %d ...", localPort)
	log.Fatal(server.ListenAndServe())
}
