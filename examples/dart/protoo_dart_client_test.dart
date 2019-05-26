import 'package:protoo_client/protoo_client.dart';

main() async {

  Peer peer = new Peer('https://127.0.0.1:8443/ws?peer=dart-client-id-xxxx');

  peer.on('open', () {

    print('open');

    peer.send('login', {"username":"alice","password":"alicespass"}).then((data) {
      print('response: ' + data.toString());
    }).catchError((error) {
      print('response error: ' + error.toString());
    });

    peer.send('offer', {'sdp':'empty'}).then((data) {
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
      if(request['method'] == 'kick')
        reject(486, 'Busy Here');
      else
        accept({});
  });

  await peer.connect();

  //peer.close();
}

