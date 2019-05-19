# go-protoo-server

## How to test.
### Clone the repository and build and run.
```
git clone https://github.com/cloudwebrtc/go-protoo-server
cd go-protoo-server
go build main.go
./main
```
### Open js example
```
cd example/js
npm i
npm start
```
### The following logs can be seen in the chrome console.
```
bundle.js:1 Peer "open" event
bundle.js:1 login success: result => {"name":"xxxx","status":"login"}
bundle.js:1 _handleRequest() [method:kick, data:Object]
bundle.js:1 offer reject: error => {"code":500}
bundle.js:1 _handleRequest() [method:kick, data:Object]
```
