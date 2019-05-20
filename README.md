# go-protoo

## How to test.
### Run go-protoo-server.
```
git clone https://github.com/cloudwebrtc/go-protoo-server
cd go-protoo-server/examples/go
go build server.go
./server
```
### Golang client test.
```
cd go-protoo-server/examples/go
go build client.go
./client
```
### JS client test.
```
cd example/js
npm i
npm start
```
### Dart client test.
```
cd example/dart
pub get
dart protoo_dart_client_test.dart

