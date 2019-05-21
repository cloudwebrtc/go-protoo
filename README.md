# go-protoo

## How to test.
### Run go-protoo-server.
```
git clone https://github.com/cloudwebrtc/go-protoo
cd go-protoo/examples/go
go build server.go
./server
```
### Golang client test.
```
cd go-protoo/examples/go
go build client.go
./client
```
### JS client test.
```
cd go-protoo/example/js
npm i
npm start
```
### Dart client test.
```
cd go-protoo/example/dart
pub get
dart protoo_dart_client_test.dart
