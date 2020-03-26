package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Usage: lncli listpayments | timingdiagram <hash> > out.html
// If hash is not given, the first payment found will be used.

func main() {
	var hash string
	if len(os.Args) == 2 {
		hash = os.Args[1]
	}

	err := paymentTiming(hash)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

type HopJson struct {
	PubKey string `json:"pub_key"`
	ChanId uint64 `json:"chan_id,string"`
}

type RouteJson struct {
	Hops     []*HopJson
	TotalAmt int `json:"total_amt,string"`
}

type FailureJson struct {
	Code               string
	FailureSourceIndex int `json:"failure_source_index"`
}

type HtlcJson struct {
	Route         *RouteJson
	AttemptTimeNs int64 `json:"attempt_time_ns,string"`
	ResolveTimeNs int64 `json:"resolve_time_ns,string"`
	Failure       *FailureJson
	Status        string
}

type PaymentJson struct {
	PaymentHash string `json:"payment_hash"`
	Htlcs       []*HtlcJson
}

type PaymentsJson struct {
	Payments []*PaymentJson
}

func paymentTiming(hash string) error {
	var data PaymentsJson
	decoder := json.NewDecoder(os.Stdin)
	err := decoder.Decode(&data)
	if err != nil {
		return err
	}

	for _, payment := range data.Payments {
		if hash != "" && payment.PaymentHash != hash {
			continue
		}

		return paymentTimingDiagram(payment)
	}

	return nil
}

func paymentTimingDiagram(payment *PaymentJson) error {
	fmt.Println(`
	<html>

	<head>
		<link href='http://fonts.googleapis.com/css?family=Open+Sans' rel='stylesheet' type='text/css'>
		<style>
			html {
				width: 10000px;
				font-family: Open Sans;
			}

			.container {
				height: 25px;
				flex-direction: row;
				display: flex;
				margin-bottom: 5px;
				align-content: center;
			}

			.htlc {
			}

			.text {
				margin-left:10px;
				display: flex;
				align-content: center;
				white-space: nowrap;
			}

			p { margin: auto; }
		</style>
	</head>

	<body>
`)

	timeStart := payment.Htlcs[0].AttemptTimeNs
	timeEnd := timeStart
	for _, h := range payment.Htlcs {
		if h.ResolveTimeNs != 0 && h.ResolveTimeNs > timeEnd {
			timeEnd = h.ResolveTimeNs + 1000000000
		}
	}

	var scale = (timeEnd - timeStart) / 1500

	for _, h := range payment.Htlcs {
		routeText := fmt.Sprintf("%v (%v) ", h.Route.Hops[0].PubKey[:6], h.Route.Hops[0].ChanId)
		for _, h := range h.Route.Hops[1:] {
			routeText += " > " + h.PubKey[:6]
		}
		text := fmt.Sprintf("%v sat (%v)", h.Route.TotalAmt, routeText)

		var start, end int64
		start = (h.AttemptTimeNs - timeStart) / scale
		if h.ResolveTimeNs != 0 {
			end = (h.ResolveTimeNs - timeStart) / scale
		} else {
			end = (timeEnd - timeStart) / scale
		}

		var background string
		switch h.Status {
		case "SUCCEEDED":
			background = "green"
		case "FAILED":
			background = "red"
			text += fmt.Sprintf(": %v @ %v", h.Failure.Code, h.Failure.FailureSourceIndex)
		default:
			background = "grey"
		}

		fmt.Printf("<div class=\"container\">")
		fmt.Printf("<div style=\"width:%vpx;\"></div>", start)
		fmt.Printf("<div class=\"htlc\" style=\"width:%vpx;background-color:%v;\"></div>", end-start, background)
		fmt.Printf("<div class=\"text\"><p>%v</p></div></div>\n", text)
	}

	fmt.Println(`
		</body>

		</html>`)

	return nil
}
