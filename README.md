consulapi
------------------------------------

`consulapi` is a simplified implementation of the http api intended to 
support Consul Connect Native while at the same time minimizing dependencies.

This repo is under heavy construction.
 
  
### Server

```go
import (
  "net"
	
  "github.com/savaki/consulapi"
  "github.com/savaki/consulapi/connect"
)

func main() {
  port := 8080

  // bind listener to port
  listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    log.Fatalln(err)
  }
  defer listener.Close()
	
  // register as connect native service
  agent := consulapi.NewAgent()
  service, err := connect.NewService(agent, "my-service", port)
  if err != nil {
  	log.Fatalln(err)
  }
  defer service.Close()
  
  // accept
  go acceptLoop(listener)
}
```

### Client

```go
import (
  "github.com/savaki/consulapi"
  "github.com/savaki/consulapi/connect"
  "google.golang.org/grpc"
  "google.golang.org/grpc/connectivity"
)

func main() {
  client := consulapi.NewHealth()
  resolver := connect.NewResolver(client, "my-service")
  conn, err := grpc.Dial("",
    grpc.WithBalancer(grpc.RoundRobin(resolver)),
    grpc.WithInsecure(),
  )
  if err != nil {
    log.Fatalln(err)
  }
  defer conn.Close()
  
  // create grpc client
}
```