# API Server

API server lisetn on "0.0.0.0:8081" by default you can modify this by providing flag values
api-host and api-port

### API
Run 
```
curl http://0.0.0.0:8081/address/0xd21934ed8eaf27a67f0a70042af50a1d6d195e81?limit=10&sortBy=time&order=asc"
``` 

API has following options
#### offset Default 0
start the records from given offset

#### limit Default 10
limit the number of records to <= limit

#### sortBy Default amount
sort by column support amount and time

#### order default
sort by order supported asc and desc


### API response
API response codes

#### Status 200
Records found returns an array of transaction information
```json
[
  {
    "from": "0x1cea940ca15a303a0e01b7f8589f39ff34308db2",
    "to": "0xd21934ed8eaf27a67f0a70042af50a1d6d195e81",
    "hash": "0xc5b324f529e87d093c908e893ff88c76f858faba424f990dec032a3fdc8c3a6b",
    "block": 13238585,
    "value": 100000000000000000,
    "fee": 846518,
    "time": "2019-09-02T21:18:24Z"
  },
  {
    "from": "0xd21934ed8eaf27a67f0a70042af50a1d6d195e81",
    "to": "0x003bbce1eac59b406dd0e143e856542df3659075",
    "hash": "0x3f92548fd0b404e187a50e91f60bb40375f4cf99ad2b116cf222b667d0197517",
    "block": 13238321,
    "value": 3000000000000000000,
    "fee": 21000,
    "time": "2019-09-02T21:00:48Z"
  }
]
```

#### Status Code 404
If server cannot find any records for the given address it returns empty response and status code 404