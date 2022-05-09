//Add handlers to each button so that events are sent to the backend.

$(function() {
    buttons = $(".btn");
    var applyBtn = buttons[0]
    var runBtn = buttons[1]
    applyBtn.onclick = function() {
        sendSettings()
        send("Apply", applyBtn.id);
        applyBtn.disabled = true
        setTimeout(() => { applyBtn.disabled = false }, 1000);
    }
    runBtn.onclick = function() {
        send("Run", runBtn.id);
    }
})

// Creates a websocket.
let ws = new WebSocket("ws://localhost:" + global.backendPort + "/web/app/events");

//Handle messages 
ws.onmessage = (message) => {
    let obj = JSON.parse(message.data);
    eventType = obj.event;

    if(eventType == "VarChange")
    {
        window[obj.VarName] = obj.Value
    }

    if(eventType == "uiUpdate")
    {
        ParseUIUpdateMessage(message);
    }

    if(eventType == "initLocations")
    {
        ParseLocationMessage(message);
    }

    if(eventType == "simulationState")
    {
        console.log(obj.Value);
        if (obj.Value) {
            $('#Run')[0].disabled = true;
        }
        else {
            $('#Run')[0].disabled = false;
        }
    }

    if(eventType == "simulationDone")
    {
        resetSketch();
    }
}

// When window is opened a message is sent.
ws.onopen = function(event)
{
    send("Ready", "Hello, World!");
}
// Sends an event with a value to a websocket. 
function send(eventName, value)
{
    console.log(eventName, value);
    ws.send(JSON.stringify({
        "Event":eventName,
        "Value": [value],}));
}

// Sends an event with an array of values to a websocket. 
function sendArray(eventName, arr)
{
    ws.send(JSON.stringify({
        "Event":eventName,
        "Value": arr,}));
}

function sendSettings(){
    console.log("sendSettings")
    sendMessageSettings(["numberOfLocations", "numberOfPersons", 
                        "spawnInfectionRisk", "spreadRisk", "simulationLimit",
                        "maskUsage", "maskEffectiveness",
                        "vaccineUsage", "vaccineEffectiveness"]);
}

// When send-settings-button is clicked this function is called.
// Saves information from the settings-form in an array and sends it through the websocket.
function sendMessageSettings(lst)
{
    var messages = [];
    for (i = 0; i < lst.length; i++)
    {
        id = lst[i];
        messages.push($('#' + id).val());
    }

    console.log(messages);

    sendArray("sendSettings", messages);
}