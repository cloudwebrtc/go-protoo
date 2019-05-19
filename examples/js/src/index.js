import {
    EventEmitter
} from 'events';
import protooClient from 'protoo-client';

export default class ProtooClientTest extends EventEmitter {

    constructor(url) {
        super()
        let transport = new protooClient.WebSocketTransport(url);
        // protoo-client Peer instance.
        this._protooPeer = new protooClient.Peer(transport);
        this._protooPeer.on('open', () => {
            console.log('Peer "open" event');
            this._login();
        });

        this._protooPeer.on('disconnected', () => {
            console.log('protoo Peer "disconnected" event');
        });

        this._protooPeer.on('close', () => {
            console.log('protoo Peer "close" event');
        });

        this._protooPeer.on('request', this._handleRequest.bind(this));

    }

    _login() {
        this._protooPeer.request('login', {}).then((data) => {
            console.log('login got success: result => ' + JSON.stringify(data));
            this._offer();
        }).catch((error) => {
            console.log('login got reject: error =>' + JSON.stringify(error));
        });
    }

    _offer() {
        this._protooPeer.request('offer', {
            'sdp': 'empty!'
        }).then((data) => {
            console.log('offer success: result => ' + JSON.stringify(data));
        }).catch((error) => {
            console.log('offer got reject: error => ' + JSON.stringify(error));
        });
    }

    _handleRequest(request, accept, reject) {
        console.log('_handleRequest() [method:%s, data:%o]', request.method, request.data);
        switch (request.method) {
            case 'kick': {
                reject(486, 'Busy Here')
            }
            break;
        default:
            console.log('other method =' + request.method);
        }
    }
}

window.ProtooClientTest = ProtooClientTest