let pc1;
let pc2;

const offerOptions = {
    offerToReceiveAudio: 1,
    offerToReceiveVideo: 1
};

// const servers = null;

const servers = {
    "iceServers": [{
        "urls": "stun:127.0.0.1:3478" //"stun:www.nvda.ren:3478" //stun:221.226.179.203:3478
    }]
};

function readyStream(stream, pc) {
    if (stream == null) {
        return;
    }

    const videoTracks = stream.getVideoTracks();
    const audioTracks = stream.getAudioTracks();
    if (videoTracks.length > 0) {
        console.log(`Using video device: ${videoTracks[0].label}`);
    }
    if (audioTracks.length > 0) {
        console.log(`Using audio device: ${audioTracks[0].label}`);
    }
    stream.getTracks().forEach(track => pc.addTrack(track, stream));
    console.log('Added local stream to pc1');
}

function initPeerConnection(callback) {
    pc1 = new RTCPeerConnection(servers);
    console.log('Created local peer connection object pc1');

    pc1.onicecandidate = e => onIceCandidate(pc1, e);
    pc1.oniceconnectionstatechange = e => onIceStateChange(pc1, e);
    pc1.ontrack = gotRemoteStream;

    if (callback) {
        callback();
    }
}

function call(stream) {
    console.log('Starting call');
    startTime = window.performance.now();

    initConnection();
    readyStream(stream, pc1);

    console.log('pc1 createOffer start');
    pc1.createOffer(onCreateOfferSuccess, onCreateSessionDescriptionError, offerOptions);
}


function onCreateSessionDescriptionError(error) {
    console.log(`Failed to create session description: ${error.toString()}`);
}

function onCreateOfferSuccess(desc) {
    console.log(`Offer from pc1
  ${desc.sdp}`);
    console.log('pc1 setLocalDescription start');
    pc1.setLocalDescription(desc, () => onSetLocalSuccess(pc1), onSetSessionDescriptionError);
    console.log('pc2 setRemoteDescription start');
    pc2.setRemoteDescription(desc, () => onSetRemoteSuccess(pc2), onSetSessionDescriptionError);
    console.log('pc2 createAnswer start');
    // Since the 'remote' side has no media stream we need
    // to pass in the right constraints in order for it to
    // accept the incoming offer of audio and video.
    pc2.createAnswer(onCreateAnswerSuccess, onCreateSessionDescriptionError);
}

function onSetLocalSuccess(pc) {
    console.log(`${getName(pc)} setLocalDescription complete`);
}

function onSetRemoteSuccess(pc) {
    console.log(`${getName(pc)} setRemoteDescription complete`);
}

function onSetSessionDescriptionError(error) {
    console.log(`Failed to set session description: ${error.toString()}`);
}

function gotRemoteStream(event) {
    if (rightVideo.srcObject !== event.streams[0]) {
        rightVideo.srcObject = event.streams[0];
        console.log('pc2 received remote stream', event);
    }
}

function onCreateAnswerSuccess(desc) {
    console.log(`Answer from pc2:
  ${desc.sdp}`);
    console.log('pc2 setLocalDescription start');
    pc2.setLocalDescription(desc, () => onSetLocalSuccess(pc2), onSetSessionDescriptionError);
    console.log('pc1 setRemoteDescription start');
    pc1.setRemoteDescription(desc, () => onSetRemoteSuccess(pc1), onSetSessionDescriptionError);
}

function onIceCandidate(pc, event) {
    getOtherPc(pc).addIceCandidate(event.candidate)
        .then(
            () => onAddIceCandidateSuccess(pc),
            err => onAddIceCandidateError(pc, err)
        );
    console.log(`${getName(pc)} ICE candidate: 
  ${event.candidate ?
            event.candidate.candidate : '(null)'}`);
}

function onAddIceCandidateSuccess(pc) {
    console.log(`${getName(pc)} addIceCandidate success`);
}

function onAddIceCandidateError(pc, error) {
    console.log(`${getName(pc)} failed to add ICE Candidate: ${error.toString()}`);
}

function onIceStateChange(pc, event) {
    if (pc) {
        console.log(`${getName(pc)} ICE state: ${pc.iceConnectionState}`);
        console.log('ICE state change event: ', event);
    }
}

function getName(pc) {
    return (pc === pc1) ? 'pc1' : 'pc2';
}

function getOtherPc(pc) {
    return (pc === pc1) ? pc2 : pc1;
}
