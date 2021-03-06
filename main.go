package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	totalRequest = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_total_request_size",
			Help: "total http request size",
		}, []string{"status", "service_name"})

	responseTimeInMs = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_time_milliseconds",
			Help: "Request completed time in milliseconds",
		}, []string{"method", "service_name", "status", "method_type", "consumer_name"})

	cusumerRequestTimes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "consumer_request_service_times",
			Help: "Request completed time in milliseconds",
		}, []string{"service_name",  "consumer_name"})
)

func init() {
	prometheus.MustRegister(totalRequest)
	prometheus.MustRegister(responseTimeInMs)
	prometheus.MustRegister(cusumerRequestTimes)
}

type KongMetrics struct {
	status int
	time   int
	module string
	method string
}

type Consumer struct {
	CreatedAt int64  `json:"created_at"`
	Username  string `json:"username"`
	Id        string `json:"id"`
}

type API struct {
	Name string `json:"name"`
}

type Request struct {
	Uri    string `json:"uri"`
	Method string `json:"method"`
}

type Response struct {
	Status int `json:"status"`
}

type Latencies struct {
	Proxy   int `json:"proxy"`
	Kong    int `json:"kong"`
	Request int `json:"request"`
}

type KongLog struct {
	Request   Request   `json:"request"`
	Response  Response  `json:"response"`
	Api       API       `json:"service"`
	Consumer  Consumer  `json:"consumer"`
	Latencies Latencies `json:"latencies"`
	ClientIp  string    `json:"client_ip"`
}

func handleKong(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var kongLog KongLog
	err := decoder.Decode(&kongLog)
	if err != nil {
		log.Println(err)
		log.Printf("handleKong Decode error\n")
		return
	}
	defer req.Body.Close()

	log.Printf("%#v\n", kongLog.Request)
	log.Printf("%#v\n", kongLog.Response)
	log.Printf("%#v\n", kongLog.Api)
	log.Printf("%#v\n", kongLog.Consumer)
	log.Printf("%#v\n", kongLog.Latencies)
	log.Printf("%#v\n", kongLog.ClientIp)

	method := kongLog.Request.Uri
	module := kongLog.Api.Name
	status := fmt.Sprint(kongLog.Response.Status)
	methodType := kongLog.Request.Method
	consumer_name := kongLog.Consumer.Username
	responseTimeInMs.With(prometheus.Labels{"method": method, "service_name": module, "status": status, "method_type": methodType, "consumer_name": consumer_name}).Observe(float64(kongLog.Latencies.Request))
	totalRequest.With(prometheus.Labels{"status": status, "service_name": module})
    cusumerRequestTimes.With(prometheus.Labels{"service_name":module,  "consumer_name":consumer_name}).Observe(float64(kongLog.Latencies.Request))
	return
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/kong", http.HandlerFunc(handleKong))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
