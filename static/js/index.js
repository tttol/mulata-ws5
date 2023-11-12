const ws = new WebSocket("ws://localhost:3001/ws");
let mediaRecorder;
let audioChunks = [];

ws.onopen = () => {
    console.log("Connected to the WebSocket server");
};

ws.onerror = (error) => {
    console.error("WebSocket Error: ", error);
};  

document.querySelector("#startRecord").addEventListener("click", function () {
    navigator.mediaDevices.getUserMedia({ audio: true })
        .then(stream => {
            mediaRecorder = new MediaRecorder(stream);
            audioChunks = [];

            mediaRecorder.ondataavailable = event => {
                ws.send(event.data);

                // WebSocket.OPEN: 1
                if (ws.readyState === 1) {
                    ws.send(event.data);
                    console.log("send audio data");
                } else {
                    console.log("Failed to send audio data.");
                }
            };

            // Start recording and generate data every 5 seconds
            // mediaRecorder.start(5000);
            startRecording();
            getTranslateResult();
        }); 
});

let isRecording = false;

const startRecording = () => {
    if (isRecording) return;
    
    console.log("start recording");
    mediaRecorder.start();
    isRecording = true;
    setTimeout(stopRecording, 5000);
}

const stopRecording = () => {
    if (!isRecording) return;

    console.log("stop recording");
    mediaRecorder.stop();
    isRecording = false;
    setTimeout(startRecording, 5000);
}

let prevText = "";

const getTranslateResult = () => {
    // if (!isRecording) return;

    fetch("http://localhost:3002/get/translate")
        .then(response => {
            console.log(`API ressponse: ${response}`);
            return response.json()
        })
        .then(data => {
            if (prevText === data.TranslatedText) return;
            document.querySelector("#result").innerHTML += "<div>" + data.TranslatedText + "</div>";
            prevText = data.TranslatedText;
        });
    
    setTimeout(getTranslateResult, 10000);
    setTimeout(getTranscribeResult, 10000);
}

const getTranscribeResult = () => {
    // if (!isRecording) return;

    fetch("http://localhost:3003/get/transcribe")
        .then(response => {
            console.log(`API ressponse: ${response}`);
            return response.json()
        })
        .then(data => {
            document.querySelector("#transcribe").innerHTML += "<div>" + data.results.transcripts[0].transcript + "</div>";
        });
    
    setTimeout(getTranslateResult, 10000);
}

document.querySelector("#closeWs").addEventListener("click", function () {
    mediaRecorder.stop();
    // navigator.mediaDevices.getUserMedia({ audio: true }).getTracks().forEach(track => track.stop());
    ws.close();
    isRecording = false;
});