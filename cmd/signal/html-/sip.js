require('./ice')

const sigServer = 'ws://127.0.0.1:55070/ws';
const ROOM_ID = 1000;
const CLIENT_ID = 10000;

const {
    CMD_REGSITER = "register",
    CMD_SEND = "send",
};

var connection = new WebSocket(sigServer); // 'ws://192.168.88.101:5535'
connection.onopen = function () {
    console.log('Connected.');

    initPeerConnection();

    initVideoStream();

};

connection.onmessage = function (message) {
    console.log('Got message', message.data);
    var data = JSON.parse(message.data);

    switch (data.eventName) {
    }
}

function register() {
    var msg = {
        cmd: CMD_REGSITER,
        roomid: ROOM_ID,
        clientid: CLIENT_ID,
        msg:''
    }
}

function send(message) {
    connection.send(JSON.stringify(message));
    console.log("send", JSON.stringify(message));
}