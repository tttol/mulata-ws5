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
            setTimeout(getTranslateResult, 3000);

            document.querySelector("#startRecord").disabled = true;
            document.querySelector("#stopRecord").disabled = false;
            document.querySelector("#closeWs").disabled = false;
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

const getTranslateResult = () => {
    fetch("http://localhost:3002/get/translate")
        .then(response => response.json())
        .then(data => {
            console.log(data);
            document.querySelector("#result").innerHTML += data.result;
        });
}

document.querySelector("#closeWs").addEventListener("click", function () {
    mediaRecorder.stop();
    ws.close();
    document.querySelector("#startRecord").disabled = false;
    document.querySelector("#closeWs").disabled = true;
});