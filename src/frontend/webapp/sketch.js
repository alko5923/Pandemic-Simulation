// Global variables 
let locations = []; 
let people = [];
let locationSize = 70;
let margin = 30;
let msgCount = -1;
let recievedMessage = false;
let personSize = 10;
let nameSize = 30;
let locnames = null;
let welcomeMsg = "Welcome to Go, Virus, Go! This is a simulation of how a virus spreads in a \
population and how the various factors can influence the rate of infection."
let instructionsMsg1 = "Let's walk you through some basic instructions to help you get started. \
At the upper edge of the screen, you have the parameter dropdown windows. \
There you can change the settings of the simulation, choosing from a variety of options, \
such as the number of locations, persons, length of simulation, etc."
let instructionsMsg2 = "Once satisfied with your choice, you apply the chosen settings \
and run the simulation by clicking the Apply and Run buttons."
let instructionsMsg3 = "On the right side of the screen you will find \
stats that are being tracked during the simulation. You can switch between \
different types of stats with the three highlighted buttons."
let instructionsMsg4 = "During the simulation, every circle represents one person: \
the blue circles are healthy, while the red circles have been infected. If a circle has \
a little white dot inside, the person has been vaccinated. If a circle has a white line \
it means the person is wearing a mask."
let instructionsMsg5 = "To re-read the instructions, click the Help button again! Enjoy!"

let nextMsg = "Next"
let closeMsg = "Close"

// *****************************************************************************


// Resets the canvas and values before a new simulation run.
function resetSketch() {
    locations = [];
    people = [];
    locationSize = 70;
    margin = 30;
    msgCount = -1;
    recievedMessage = false;
    personSize = 10;
    nameSize = 30;
    locnames = null;
}

// Displays a simple welcome message.
function welcomeUser() {
    nextTick = 0;
    $("#welcome-holder").show();
    $("#welcome-content").html(welcomeMsg);
    $("#next").html(nextMsg);
    $("#close").html(closeMsg);
}

let infectAmount = [];
let infectRateAmount = [];

let img1 = null;
let img2 = null;

let simWidth = 800;
let simHeight = 800;

let timestep = 0;

let statSelection = 0;

//Setup the p5 environment.
function setup()
{
    createCanvas(1600, 800);

    img1 = createGraphics(simWidth,simHeight);
    img2 = createGraphics(width - simWidth - 60, height);
}

//Change what statswindow should be drawn.
function SetStatWindow(idx)
{
    statSelection = idx;
}

//Draw the p5 elements
function draw()
{   
    img1.background('grey');
    img2.background('grey');


    for (let i = 0; i < locations.length; i++)
    {
        locations[i].show();
    }

    for(let i = 0; i < people.length; i++){
        people[i].show();
        people[i].update();
    }

    let done = false;

    // check if simulation for this timeTick is done
    for(let i = 0; i < people.length; i++){
        done = people[i].target == null;
        showLocStats();
    }

    // send message to backend 
    if(done && recievedMessage)
    {
        recievedMessage = false;
        send("timeSynchUI", "done");
    }
    
    img2.textSize(12);
    

    switch(statSelection)
    {
        case 0:
            DrawInfectedGraph();
            break;
        case 1:
            DrawInfectRateGraph();
            break;
        case 2:
            DrawInfectedStats();
            break;
        default:

            break;
    }

    image(img1, 0, 0);
    image(img2, simWidth + 10, 0);
}

//Draw a graph that shows the number of infected people
function DrawInfectedGraph()
{
    DrawZoom();

    let trueMaxValue = max(infectAmount);
    let maxValue = trueMaxValue * 1.35;

    if(maxValue == 0 || maxValue == null)
        return; 

    let modifier = 50 / scale;

    for (let i = 0; i < maxValue; i += scale * 4)
    {
        img2.text(i, 50, 700 - i * modifier / 4);
    }

    for(let i = 0; i < timestep; i += 5)
    {
        img2.text(i, i * 5 + 85, 730);

        let barHeigth = map(infectAmount[i], 0, trueMaxValue, 0, trueMaxValue * modifier / 4);

        img2.rect(i * 5 + 85, 700, 20, -barHeigth);
        img2.text(infectAmount[i], i * 5 + 85, 700 - barHeigth - 10)
    }
}

//Handle mouse presses.
function mousePressed()
{
    let offsetX = mouseX - img1.width - 10;

    if(offsetX >= 5 && offsetX <= 69 && mouseY >= 5 && mouseY <= 69)
    {
        zoom(0);
    }

    if(offsetX >= 75 && offsetX <= 69 + 70 && mouseY >= 5 && mouseY <= 69)
    {
        zoom(1);
    }
}

let scale = 25;

//Change the zoom values.
function zoom(dir)
{
    if(dir == 0)
    {
        scale -= 5
        if(scale <= 10)
            scale = 10;
    }
    else
    {
        scale += 5
    }
}

//Draw the zoom controlls.
function DrawZoom()
{
    img2.textSize(32)
    img2.rect(5, 5, 64, 64);
    img2.rect(75, 5, 64, 64);

    img2.text("+", 25, 50)
    img2.text("-", 97, 50)

    img2.textSize(12)
}

//Draw a graph that shows the number of infected people per timestep.
function DrawInfectRateGraph()
{
    DrawZoom();

    let localRateArr = [];

    for(let i = 0; i < infectRateAmount.length; i += 5)
    {
        let sum = 0;
        for(let j = 0; j < 5; j++)
        {
            let val = infectRateAmount[i + j];

            sum += val == undefined ? 0 : val;
        }

        localRateArr.push(sum);
    }

    let trueMaxValue = max(localRateArr);
    let maxValue = trueMaxValue * 1.35;

    if(maxValue == 0 || maxValue == null)
        return;

    
    let modifier = 50 / scale;
    
    for (let i = 0; i < maxValue; i += scale)
    {
        img2.text(i, 50, 700 - i * modifier);
    }

    for(let i = 0; i < timestep; i += 5)
    {
        img2.text(i, i * 5 + 85, 730);

        let barHeigth = map(localRateArr[i / 5], 0, trueMaxValue, 0, trueMaxValue * modifier);

        img2.rect(i * 5 + 85, 700, 20, -barHeigth);
        img2.text(localRateArr[i / 5], i * 5 + 85, 700 - barHeigth - 10)
    }
}

//Draw basic stats.
function DrawInfectedStats()
{
    let str = "";
    str += "Locations: " + locations.length + "\n";
    str += "People: " + people.length + "\n";
    str += "Infected: " + getInfected() + " / " + people.length + "\n";
    str += "Vaccinated: " + getVaccinated() + " / " + people.length + "\n";
    str += "Masks: " + getMasked() + " / " + people.length + "\n";

    img2.textSize(32);
    img2.text(str, 10, 80);
}


// Shows stats for locations that the mouse hovers over 
function showLocStats() 
{   
    if(statSelection != 2) {
        return;
    }

    for (let i = 0; i < locations.length; i++) {
        let loc = locations[i];
        let locPopulation = 0;
        let infectedPopulation = 0;
        let maskedPopulation = 0;
        let vaccinatedPopulation = 0;
        // Check if mouseX and mouseY is within location limits
        if (mouseX >= loc.x && mouseX <= loc.x+loc.size) {
            if (mouseY >= loc.y && mouseY <= loc.y+loc.size) {
                let stats = "STATS FOR CHOSEN LOCATION";
                img2.textSize(20);
                img2.text(stats, 10, 300);
                // Loop through list of persons and check for persons 
                // that are on the location we want to show statistics for.
                for (let j = 0; j < people.length; j++) {
                    if (loc.x == people[j].location.x && loc.y == people[j].location.y && people[j].done==false) {
                        locPopulation++;
                        if (people[j].infected) infectedPopulation++;
                        if (people[j].masked) maskedPopulation++;
                        if (people[j].vaccinated) vaccinatedPopulation++;
                    }
                }
                img2.text("Visitors: " + locPopulation, 10, 340);
                img2.text("Infection scale: " + infectedPopulation + " / " + locPopulation, 10, 380);
                img2.text("Persons wearing mask: " + maskedPopulation + " / " + locPopulation, 10, 420);
                img2.text("People vaccinated: " + vaccinatedPopulation + " / " + locPopulation, 10, 460);
            }
        }
    }
}

//Reset the p5 environment and entire sketch.
function resetSketch() 
{
    locations = [];
    people = [];
    locationSize = 70;
    margin = 30;
    msgCount = -1;
    recievedMessage = false;
    personSize = 10;
    nameSize = 30;
    locnames = null;
    infectAmount = [];
    infectRateAmount = [];
    timestep = 0;
}

// Walks the user through the basic instructions 
function showInstructions() {
    if (nextTick == 0) {
        $("#welcome-content").html(instructionsMsg1);
        highlightDiv("#population");
        highlightDiv("#virus");
        highlightDiv("#protection");
    } else if (nextTick == 1) {
        $("#welcome-content").html(instructionsMsg2);
        unhighlightDiv("#population");
        unhighlightDiv("#virus");
        unhighlightDiv("#protection");
        highlightDiv("#Apply");
        highlightDiv("#Run");
    } else if (nextTick == 2) {
        $("#welcome-content").html(instructionsMsg3);
        unhighlightDiv("#Apply");
        unhighlightDiv("#Run");
        highlightDiv("#stats1");
        highlightDiv("#stats2");
        highlightDiv("#stats3");
    } else if (nextTick == 3) {
        $("#welcome-content").html(instructionsMsg4);
        unhighlightDiv("#stats1");
        unhighlightDiv("#stats2");
        unhighlightDiv("#stats3");
    } else if (nextTick == 4) {
        $("#welcome-content").html(instructionsMsg5);
        highlightDiv("#help")
    } else if (nextTick == 5) {
        unhighlightDiv("#help");
        closeWelcome();
    }
    nextTick++;
}

// Highlights chosen elements during instructions phase. 
function highlightDiv(divID) {
    let addclass = 'color';
    $(divID).addClass(addclass);
}

// Removes the highlight from chosen elements during instructions phase. 
function unhighlightDiv(divID) {
    let removeclass = 'color';
    $(divID).removeClass(removeclass);
}

// Closes the welcome window. 
function closeWelcome() {
    $("#welcome-holder").hide();
    resetAllHighlights();
}

// Resets all highlighted elements if the instructions runthrough is closen before finished. 
function resetAllHighlights() {
    var elements = document.getElementsByClassName('color');
    let ids = [];
    for(let i=0; i<elements.length; i++) {
        let id = elements[i].getAttribute("id");
        let idName = "#" + id;
        ids.push(idName)
    }
    for(let i=0; i<ids.length; i++) {
        unhighlightDiv(ids[i]);
    }
}

// Creates locations.
function CreateLocations(amountLocs)
{
    locations = [];
    let choices = [];

    // Generate list of positions that are valid and don't intersect
    for(let x = 0; x < simWidth; x += locationSize + margin)
    {
        for(let y = locationSize + margin; y < simHeight; y += locationSize + margin)
        {
            choices.push({x, y});
        }
    }

    // Generate new locations.
    for(let i = 0; i < amountLocs; i++)
    {
        // Pick a random position for the location.
        let choice = random(choices);

        // Remove the position so that it cannot be picked again.
        let idx = choices.indexOf(choice);
        choices.splice(idx, 1);

        let x = choice.x;
        let y = choice.y;

        // Make the location.
        let loc = new Location(x, y, locnames == null ? "Loc " + i : locnames[i]);

        // Add it to the array.
        locations.push(loc);
    }
}

// Parses the message from the backend containing information about locations 
function ParseLocationMessage(message)
{
    let json = JSON.parse(message.data).Value;
    locnames = json.Names;

    for(let i = 0; i < locations.length; i++)
    {
        locations[i].name.name = locnames[i];
    }
}

// Parses the message from the backend containing information about updated state after a time step
function ParseUIUpdateMessage(message)
{
    recievedMessage = true;
    msgCount++;
    timestep++;

    if(msgCount == 0)
    {
        Init(message);
        return;
    }
    
    let value = JSON.parse(message.data).Value;

    if(value == null)
    {
        infected = getInfected();
        infectAmount.push(infected);
        return;
    }

    let oldInfected = getInfected();

    // Goes through persons and checks their updated status
    for(let i = 0; i < value.length; i++)
    {
        let person = value[i];

        people[person.PersonID].infected = person.Infected;
        // Value of -1 means that they haven't moved
        if(person.CurrentLocation >= 0)
        // Set a new target for the persons that need to be moved 
            people[person.PersonID].setTargetByIdx(person.CurrentLocation);
        if (person.CurrentLocation === -2)
            // Handle the persons that have finished their path 
            people[person.PersonID].done = true
    }

    infected = getInfected();
    infectAmount.push(infected);
    infectRateAmount.push(infected - oldInfected);
}

// Initializes the visuals: creates locations and populates them with persons.   
function Init(message)
{
    let locationAmount = parseInt($("#numberOfLocations").val());
    CreateLocations(locationAmount);

    let value = JSON.parse(message.data).Value;

    people = new Array(value.length);
    creatingPeople = true;
    for(let i = 0; i < value.length; i++)
    {
        let person = value[i];
        let loc = locations[person.CurrentLocation];
        let startPos = loc.randomPosWithin();

        let p = new Person(person.PersonID, startPos, loc, person.Infected, null, 2, person.Masked, person.Vaccinated);
        people[p.idx] = p;
    }

    infected = getInfected();
    infectAmount.push(infected);
    infectRateAmount.push(infected);
}

// Returns the amount of people infected. 
function getInfected()
{
    let counter = 0;
    
    for(let i = 0; i < people.length; i++)
    {
        if(people[i].infected)
            counter++;
    }
    
    return counter;
}

// Returns the amount of people wearing masks. 
function getMasked() {
    let counter = 0;
    for (let i = 0; i < people.length; i++) {
        if (people[i].masked)
            counter++;
    }
    return counter;
}

// Returns the amount of people who have been vaccinated.
function getVaccinated() {
    let counter = 0;
    for (let i = 0; i < people.length; i++) {
        if (people[i].vaccinated)
            counter++;
    }
    return counter;
}

// creates class Name which refers to the names of locations
class Name 
{
    constructor(x,y,name) {
        this.x = x;
        this.y = y;
        this.name = name;
        this.size = nameSize;
    }
    show() {
        img1.fill(0,0,0);
        let firstPart = "";
        let secondPart = "";
        img1.textStyle(ITALIC);
        let index = this.name.indexOf(" ");

        if (this.name.length > 11) 
        {
            if(this.name.includes(" ")) {
                firstPart = this.name.slice(0,index);
                secondPart = this.name.slice(index+1);
                img1.noStroke();
                img1.text(firstPart, this.x+2, this.y-12);
                img1.text(secondPart, this.x+2, this.y-2);
                img1.stroke(color('black'));
            }
            else {
                img1.noStroke();
                img1.text(this.name, this.x-1, this.y-2);
                img1.stroke(color('black'));
            }
        }
        else if (this.name.length > 8) {
            img1.noStroke();
            img1.text(this.name, this.x+4, this.y-2);
            img1.stroke(color('black'));
        }
        else {
            img1.noStroke();
            img1.text(this.name, this.x+8, this.y-2);
            img1.stroke(color('black'));
        }
    }
}

// Creates class Location. 
class Location 
{
    constructor(x,y,name) {
        this.x = x;
        this.y = y;
        this.name = new Name(x, y,name);
        this.size = locationSize;
    }

    // Displays a location. 
    show() 
    {
        img1.fill(0,0,0);
        img1.square(this.x,this.y,this.size);
        this.name.show();
    }

    // Finds the coordinates of the location's center. 
    getCenter(){
        return createVector(this.x + this.size / 2, this.y + this.size /2);
    }

    // Returns a random position within the bounds of the location. 
    randomPosWithin()
    {
        let x = this.getCenter().x + random(-this.size / 2 + personSize / 2, this.size / 2 - personSize / 2);
        let y = this.getCenter().y + random(-this.size / 2 + personSize / 2, this.size / 2 - personSize / 2);

        return createVector(x, y);
    }
}

//Creates class Person.
class Person
{
    constructor(idx, startPos, location, infected, target, speed, masked, vaccinated) {
        this.idx = idx;
        this.startPos = null;
        this.infected = infected;
        this.masked = masked;
        this.vaccinated = vaccinated;
        this.x = startPos.x;
        this.y = startPos.y;
        this.location = location;
        this.target = target;
        this.oldTarget = createVector(this.x, this.y);
        this.percentage = 0;
        this.speed = speed;
        this.size = personSize;
        this.done = false
        this.shouldRoam = true;
    }

    // Sets the status of the person to infected. 
    setInfectedStatus(infected)
    {
        this.infected = infected;
    }

    // Sets the target the person will be moved to. 
    setTarget(target)
    {
        if(this.target != null)
            this.oldTarget = this.target;
        this.target = target;
        this.percentage = 0;
    }

    // Sets the target to the appropriate location. 
    setTargetByIdx(idx)
    {
        let newTarget = locations[idx].randomPosWithin();
        this.location = locations[idx];
        this.shouldRoam = false;
        this.setTarget(newTarget);
    }

    // Sets the color of the person, red for infected and blue for not infected, 
    // and draws the person. 
    show()
    {
        if (this.done) {
            return
        }
        let infectedColour = color(255,0,0);
        let nonInfectedColour = color(0,0,255);
        img1.fill(this.infected ? infectedColour : nonInfectedColour);
        img1.circle(this.x, this.y, this.size);
        img1.noFill();
        if (this.vaccinated) {
            img1.fill('white');
            img1.circle(this.x,this.y, this.size-6);
            img1.noFill();
        }
        if (this.masked) {
            img1.stroke(color('white'));
            img1.circle(this.x, this.y, this.size);
            img1.stroke(color('black'));
        }
    }
    
    // Makes the persons move within the boundaries of a location while they are 
    // waiting to be moved to another location. 
    roam()
    {
        this.setTarget(this.location.randomPosWithin());
    }

    // Updates the position of the person. 
    update()
    {
        if(this.target == null && this.shouldRoam)
        {
            this.roam();
        }

        if(this.target == null)
            return;

        this.percentage += 0.01 * this.speed;
        
        if(this.percentage > 1)
            this.percentage = 1;
        
        let vNew = p5.Vector.lerp(this.oldTarget, this.target, this.percentage);
        
        this.x = vNew.x;
        this.y = vNew.y;
        
        if(this.percentage >= 1)
        {
            this.oldTarget = this.target;
            this.target = null;
            this.shouldRoam = true;
        }   
    }
}

