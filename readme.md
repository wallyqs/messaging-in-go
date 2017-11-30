Based on the O'Reilly series published at:

- https://www.oreilly.com/ideas/building-messaging-in-go-network-clients
- https://www.oreilly.com/ideas/scaling-messaging-in-go-network-clients
- https://www.oreilly.com/ideas/implementing-request-response-using-context-in-go-network-clients

To run:

```
go run main.go parser.go client.go
```

Example Result:

```
2017/11/19 12:03:01 [Wildcard] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Received] world
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.f1e837571be29a83412ffa] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.397ee5efd22cfd4b6c3d2a] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.f13a93a4ecd31bec3a58c9] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.ef96dccf7b48312ca70596] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.81305ec29c6b0bc483ace1] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.002266efafc4685b3db996] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.603a1f3e615d2021fb686c] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.1e6a5ce1570766369c732e] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.42ca349d161368d055bbec] hello!
2017/11/19 12:03:01 [Response] hi!
2017/11/19 12:03:01 [Request:_INBOX.096474adbf71508188ffdb.6f817ec3f21846ffc03907] hello!
2017/11/19 12:03:02 [Wildcard] world
2017/11/19 12:03:03 [Wildcard] world
2017/11/19 12:03:04 [Wildcard] world
2017/11/19 12:03:05 [Wildcard] world
2017/11/19 12:03:06 [Wildcard] world
2017/11/19 12:03:07 [Wildcard] world
2017/11/19 12:03:08 [Wildcard] world
2017/11/19 12:03:09 [Wildcard] world
2017/11/19 12:03:10 [Wildcard] world
```

To run the included benchmark:

```
go test ./... -v -bench=Benchmark_Publish -benchmem
```
