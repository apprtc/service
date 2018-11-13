var name = 'xxx';

const leftVideo = document.getElementById('leftVideo');
let rightVideo = document.querySelector('#theirs');
let pc1;


let stream;

function login() {
    setupPeerConnection();
    send({
        eventName: 'join',
        metaData: {
            userId: name,
            roomId: 100
        }
    })
}


function maybeCreateStream() {
    console.log('maybeCreateStream');
    if (stream) {
        return;
    }
    if (leftVideo.captureStream) {
        stream = leftVideo.captureStream();
        console.log('Captured stream from leftVideo with captureStream',
            stream);
        //   call(stream);
    } else if (leftVideo.mozCaptureStream) {
        stream = leftVideo.mozCaptureStream();
        console.log('Captured stream from leftVideo with mozCaptureStream()',
            stream);
        //   call(stream);
    } else {
        console.log('captureStream() not supported');
        return;
    }

    login();
}


function initVideoStream() {
    // Video tag capture must be set up after video tracks are enumerated.
    leftVideo.oncanplay = maybeCreateStream;
    if (leftVideo.readyState >= 3) { // HAVE_FUTURE_DATA
        // Video is already ready to play, call maybeCreateStream in case oncanplay
        // fired before we registered the event handler.
        maybeCreateStream();
    }

    leftVideo.play();
}

var connection = new WebSocket('ws://127.0.0.1:55070/ws'); // 'ws://192.168.88.101:5535'
connection.onopen = function () {
    console.log('Connected.');

    initVideoStream();

};
connection.onmessage = function (message) {
    console.log('Got message', message.data);
    var data = JSON.parse(message.data);
    switch (data.eventName) {
        case 'peer':
            onPeer();
            break;
        case 'newpeer':
            onNewPeer();
            break;
        case 'bye':
            onBye();
            break;
        case 'offer':
            onOffer(data.metaData.sdp);
            break;
        case 'answer':
            onAnswer(data.metaData.sdp);
            break;
        case 'candidate':
            onCandidate(data.metaData);
            break;
        case 'leave':
            onLeave();
            break;
        default:
            break;
    }
};
connection.onerror = function (err) {
    console.log("Got error", err);
};

function send(message) {
    connection.send(JSON.stringify(message));
    console.log("send", JSON.stringify(message));
}

function onPeer() {
}

function onNewPeer() {
    startPeerConnection()
}

function onBye() {
    onLeave();
}

function onOffer(offer) {
    pc1.setRemoteDescription(new RTCSessionDescription({
        type: 'offer',
        sdp: offer
    }));
    pc1.createAnswer(function (_answer) {
        pc1.setLocalDescription(_answer);
        send({
            eventName: 'answer',
            metaData: {
                sdp: _answer.sdp
            }
        });
    }, function (err) {
        console.log('An error occur on onOffer.', err);
    });
};

function onAnswer(answer) {
    pc1.setRemoteDescription(new RTCSessionDescription(
        {
            sdp: answer,
            type: 'answer'
        }));
};

function onCandidate(candidate) {
    pc1.addIceCandidate(new RTCIceCandidate({
        candidate: candidate.sdp,
        sdpMLineIndex: candidate.index,
        sdpMid: candidate.mid
    }));

    console.log('onCandidate end')


};

function onLeave() {
    pc1.close();
    pc1.onicecandidate = null;
    pc1.onaddstream = null;
    setupPeerConnection(stream);
};

function hasUserMedia() {
    return !!(navigator.getUserMedia || navigator.webkitGetUserMedia || navigator.msGetUserMedia || navigator.mediaDevices.getUserMedia || navigator.mozGetUserMedia);
}

function hasRTCPeerConnection() {
    window.RTCPeerConnection = window.RTCPeerConnection || window.webkitRTCPeerConnection || window.mozRTCPeerConnection;
    window.RTCSessionDescription = window.RTCSessionDescription || window.webkitRTCSessionDescription || window.mozRTCSessionDescription;
    window.RTCIceCandidate = window.RTCIceCandidate || window.webkitRTCIceCandidate || window.mozRTCIceCandidate;
    return !!window.RTCPeerConnection;
}

function gotRemoteStream(event) {
    console.log('ontrack stream=', event)
    if (rightVideo.srcObject !== event.streams[0]) {
        rightVideo.srcObject = event.streams[0];
        console.log('pc1 received remote stream', event);
    }
}

function setupPeerConnection() {
    console.log('setupPeerConnection')
    var configuration = {
        "iceServers": [{
            "urls": "stun:221.226.179.203:3478"//"stun:www.nvda.ren:3478" //
        }]
    };

    if (hasRTCPeerConnection()) {
        pc1 = new RTCPeerConnection(configuration);


    }
    else {
        alert('Sorry, your browser does not support WebRTC.');
        return;
    }

    // pc1.onaddstream = function (e) {
    //     console.log('onaddstrean stream=', e)
    //     //rightVideo.src = e.stream;//window.URL.createObjectURL(e.stream);
    //     rightVideo.srcObject = e.stream;
    // };

    pc1.onicecandidate = function (e) {
        console.log('onicecandidate');
        if (e.candidate) {
            send({
                type: "candidate",
                candidate: e.candidate
            });
        }
    };


    pc1.ontrack = gotRemoteStream;
    stream.getTracks().forEach(track => {
        pc1.addTrack(track, stream)
    });
    console.log('Added local stream to pc1');

};

function startPeerConnection() {
    pc1.createOffer(function (_offer) {
        send({
            eventName: "offer",
            metaData: { sdp: _offer.sdp }
        });
        pc1.setLocalDescription(_offer);
    }, function (error) {
        alert('An error on startPeerConnection:', error);
    });
};