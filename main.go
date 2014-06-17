package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/eth-go"
	"github.com/ethereum/eth-go/ethutil"
	"github.com/ethereum/go-ethereum/utils"
	"github.com/miekg/dns"
	"strings"
)

/*
* Try it with:
* dig @localhost -p 8053 bytesized-hosting.eth
* Problem: No recursion
 */
var Datadir string
var Port string
var DnsPort string
var MaxPeer int

var dnsreg = ethutil.FromHex("1b6a704f1c12e98b4b355d385e8eeaa7e7b237e2")

func NewRR(s string) dns.RR { r, _ := dns.NewRR(s); return r }

func initEthereum() *eth.Ethereum {
	// Create flag for datadir folder
	flag.StringVar(&Datadir, "datadir", ".ethdns", "specifies the datadir to use. Takes precedence over config file.")
	flag.StringVar(&Port, "port", "20202", "specifies the Ethereum port to use.")
	flag.StringVar(&DnsPort, "dnsport", "8053", "specifies the DNS port to use.")
	flag.IntVar(&MaxPeer, "maxpeer", 10, "maximum desired peers")

	flag.Parse()

	// Set logging flags
	var lt ethutil.LoggerType
	lt = ethutil.LogFile | ethutil.LogStd

	// Read config if any
	ethutil.ReadConfig(Datadir, lt, nil, "DnsEth")

	// Create Ethereum object
	ethereum, err := eth.New(eth.CapDefault, false)

	if err != nil {
		fmt.Println("Could not start the Ethereum-core:", err)
		return nil
	}
	// Make sure we have an public key to identify ourselves
	utils.CreateKeyPair(false)

	// Set the port
	ethereum.Port = Port

	// Set the max peers
	ethereum.MaxPeers = MaxPeer

	return ethereum
}

func updateThread(ethereum *eth.Ethereum) {
	objectChan := make(chan ethutil.React, 1)
	reactor := ethereum.Reactor()
	reactor.Subscribe("object:"+string(dnsreg), objectChan)
	for {
		select {
		case <-objectChan:
			updateDns(ethereum)
		}
	}
}
func sanitizeString(val string) string {
	return strings.Trim(val, "\000 ")
}

func updateDns(ethereum *eth.Ethereum) {
	stateObject := ethereum.StateManager().CurrentState().GetStateObject(dnsreg)
	if stateObject != nil {
		ethutil.Config.Log.Debugln("Updating DNS")
		stateObject.State().EachStorage(func(name string, value *ethutil.Value) {
			val := value.Bytes()[1:]
			name = sanitizeString(name) + ".eth."
			dns.HandleRemove(name)
			zoneString := fmt.Sprintf("%s 2044 IN A %s", name, val)
			zone := NewRR(zoneString)
			if zone != nil {
				ethutil.Config.Log.Debugln("[DNS] Updated zone:", zone)
				dns.HandleFunc(name, func(w dns.ResponseWriter, r *dns.Msg) {
					switch r.Question[0].Qtype {
					case dns.TypeA:
						m := new(dns.Msg)
						m.SetReply(r)
						m.Answer = []dns.RR{zone}
						m.RecursionAvailable = true
						w.WriteMsg(m)
					default:
						ethutil.Config.Log.Debugln("[DNS] Type not supported yet")
					}

				})
			} else {
				ethutil.Config.Log.Debugln("Invalid zone", zoneString)
			}
		})
	}
}

func main() {
	ethereum := initEthereum()
	ethutil.Config.Log.Infoln("Starting DNS server on port:", DnsPort)

	server := &dns.Server{Pool: false, Addr: ":" + DnsPort, Net: "udp", TsigSecret: nil}
	go server.ListenAndServe()

	go updateThread(ethereum)

	// Set initial records
	updateDns(ethereum)

	ethereum.Start(true)

	// Wait for shutdown
	ethereum.WaitForShutdown()
}
