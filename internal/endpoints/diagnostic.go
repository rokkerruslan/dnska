package endpoints

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	successesProcessedOpsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_server_successes_processed_ops_total",
		Help: "The total number of processed resolve requests",
	})

	packetReadErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_server_packet_read_errors_total",
		Help: "The total number of read errors happen while read UDP packet from server socket",
	})

	packetDecodeErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_server_packet_decode_error_total",
		Help: "The total number of packet decode errors",
	})

	packetWriteErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_server_packet_write_error_total",
		Help: "The total number of write errors happen while writing into a client socket",
	})

	packetEncodeErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dnska_server_packet_encode_error_total",
		Help: "The total number of packet encode errors",
	})
)
