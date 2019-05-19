import 'package:protoo_client/protoo_client.dart';

main() async {

  Peer peer = new Peer('https://127.0.0.1:8443/ws?peer-id=xxxxx');

  peer.on('open', () {

    print('open');

    peer.send('login', {}).then((data) {
      print('response: ' + data.toString());
    }).catchError((error) {
      print('response error: ' + error.toString());
    });

    peer.send('offer', {'sdp':'empty!'}).then((data) {
      print('response: ' + data.toString());
    }).catchError((error) {
      print('response error: ' + error.toString());
    });
  });

  peer.on('close', () {
    print('close');
  });

  peer.on('error', (error) {
    print('error ' + error);
  });

  peer.on('request', (request, accept, reject) {
    print('request: ' + request.toString());
    accept({ 'key1':"value1", 'key2':"value2"});
    //reject(404, 'Oh no~~~~~');
  });

  await peer.connect();

  //peer.close();
}

