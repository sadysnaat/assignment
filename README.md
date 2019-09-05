### Building

#### Source Code
Run the following assuming you are running go 1.11 or above 
```bash
go get github.com/sadysnaat/assignment
```
this will install a binary assignment in your $GOBIN

#### Dockerfile
this package comes with a dockerfile 
to build the project do

docker build -t <image_name> .

this will generate the docker image


Command will require five options
```go
wssURL := flag.String("wss-url", "wss://kovan.infura.io/ws/v3/6c6f87a10e12438f8fbb7fc7c762b37c", "websocket url for the subscription")
httpsURL := flag.String("https-url", "https://kovan.infura.io/v3/6c6f87a10e12438f8fbb7fc7c762b37c", "https url for indexer")
dbURL := flag.String("db-url", "root:my-secret-pw@tcp(localhost:32768)/assignment", "database uri")
apiHost := flag.String("api-host", "0.0.0.0", "api host")
apiPort := flag.String("api-port", "8081", "api port")
``` 
if you are running from docker you will need to provide -db-url in the option 
if indexer cannot connect to db it will panic

#### Preparing the database
database should have one user configured to able to login from different ip addresses.
Docker networking on mac makes it difficult to test the connection

create a database with <database_name>
use in -db-url root:my-secret-pw@tcp(localhost:32768)/<database_name>

Run 
store/blocks.sql
store/transactions.sql
in the same order 

Database is ready to use.

#### Running the main binary
run 
```go
assignment -db-url="<user_name>:<password>@tcp(<host_name>:<port>)/<database_name>"
````
 

#### API Guide
Please refer API.md

#### Design 
Please refer DESIGN.md