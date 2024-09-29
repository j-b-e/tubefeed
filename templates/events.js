const eventSource = new EventSource('/events');

console.log("ready: ",eventSource.readyState, " ", eventSource.url)

// eventSource.onerror = (e) => {
//     console.log("error: ",e)
// }

eventSource.addEventListener("ping", (e) => {
    console.log("ready: ",eventSource.readyState, " ", eventSource.url)
    console.log("PING: ", e.data);
});

eventSource.addEventListener("newVideo", (e) => {
        console.log("newVideo: ", e);
        console.log("newVideo: ", e.data);
        data = JSON.parse(e.data)
        audioID = data.audioID
        console.log("audioID: ", audioID);
    // Add new video to the list
        const videoList = document.getElementById("video-list");
        const newVideoElement = document.createElement("div");
        newVideoElement.id = "audio-" + audioID;
        newVideoElement.innerHTML = `<span>${audioID} - Status: <span id="status-${audioID}">Pending</span></span>`;
        videoList.appendChild(newVideoElement);
});

eventSource.addEventListener("statusUpdate", (e) => {
        console.log("statusUpdate: ", e.data);
        // Update the status of an existing video
        const statusElement = document.getElementById("status-" + audioID);
        if (statusElement) {
            statusElement.innerText = status;
        }
});

eventSource.addEventListener("deleteVideo", (e) => {
        console.log("deleteVideo: ", e.data);
        data = JSON.parse(e.data)
        const videoElement = document.getElementById("audio-" + data.audioID);
        videoElement.remove()
});