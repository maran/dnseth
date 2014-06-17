# DNS Ethereum
DNSEth is a DNS server prototype that uses Ethereum as backend.

It will only ever support A records and has no recursion so make sure
you have a secondary DNS server setup in case you want to give this a
go.


## How to register a name
* Boot [Ethereal](https://github.com/ethereum/go-ethereum)
* Send a transaction to 1b6a704f1c12e98b4b355d385e8eeaa7e7b237e2
* Fill in a 0 and two strings, domain (without the .eth extension) and ip, as arguments in the data field.

For instance:
```
0
"ethereum"
"62.251.77.75"
```

## How to use the DNS
Setup 94.242.229.217 as DNS server and use 8.8.8.8 as secondary.

With this setup you can for instance go to http://maran.eth/ and if all is well it should work as expected.
