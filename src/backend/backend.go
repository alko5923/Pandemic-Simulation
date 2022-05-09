package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Equanox/gotron"
	"github.com/ospp-projects-2021/clockwork/src/data"
)

//////////////// Variables ////////////////////

// Variable to give persons an ID
var id int = 0
var idMutex sync.Mutex
// Variables for vaccine info
var vaccinePrevention int = 0
var vaccinePrevMutex sync.Mutex
var vaccineNoPrevention int = 0
var vaccineNoPrevMutex sync.Mutex
// Struct that stores simulation parameters.
var paramStruct parameters = parameters{}

// Slice of names for locations
var locationNames = [26]string{
	"ICA MAXI",
	"Apoteket",
	"Webhallen",
	"Biblioteket",
	"Uppsala Klättercentrum",
	"Clas Ohlson",
	"Pollax",
	"Palermo",
	"Ofvandahls",
	"Lutis",
	"Folkes",
	"ICA Väst",
	"Gymmet",
	"Espresso House",
	"Tropiska växthuset",
	"Nordea Bank",
	"Hemköp",
	"Evolutionsmuseet",
	"Slottet",
	"Systembolaget",
	"Stadsparken",
	"Ekoparken",
	"Botaniska",
	"Gamla Kyrkogården",
	"Fyrisån",
	"Kåbo Växthus",
}

// Slice of modifiers that alters the spread risk at each location
var locationModifier = [26]float64{
	1,
	1.7,
	1.6,
	0.5,
	1,
	0.8,
	1.3,
	2,
	1.3,
	1.5,
	1.6,
	1.4,
	1.1,
	1.3,
	0.4,
	0.5,
	1.2,
	1.4,
	0.3,
	0.9,
	0.1,
	0.1,
	0.1,
	0.1,
	0.1,
	0.4,
}
// Slice to hold the synchChannels used by the EventLoop to synchronize LocationWorkers
var synchChannels []chan Synch
// Slice to hold the doneChannels used by the EventLoop to tell LocationWorkers to terminate
var doneChannels []chan bool
// Channel used to transmit data to the UI
var transmitToUI chan []data.UpdatePersonUI = make(chan []data.UpdatePersonUI)
// Channels used to transmit statistics between various modules and the statisticshandler
var statistics statisticsChannels = statisticsChannels{
	parameters: make(chan string),
	infection:  make(chan logInfo, 10),
	location:   make(chan logInfo, 10),
	person:     make(chan logInfo, 10),
	done:       make(chan int),
	exit:		make(chan bool),
}
// Slice for holding all locations that are created
var allLocations []location

//////////////// Data structures ////////////////////
// Struct representing a single person
type Person struct {
	personID         int   // the person's ID
	infected         bool  // if the person is infected
	infectedTime     int   // when the person was infected
	infectedLocation int   // where the person was infected
	path             []int // which locations the person will visit
	departureTime    int   // when the person is leaving the current location
	departureTimes   []int // list of departure times 
	nextStop         int   // where the person will go next
	locationsVisited int   // how many locations the person has visited
	done             bool  // if the person has finished travelling
	vaccinated       bool  // if the person is vaccinated
	masked           bool  // if the person wears a mask
}

// Struct used for time synchronization
type Synch struct {
	time      int              // The current time step
	waitGroup *sync.WaitGroup  // WaitGroup for current time step
}

// Struct representing a single location
type location struct {
	Name      string       // location name
	receiveCh chan Person  // channel for receiving persons 
	modifier  float64      // spread risk modifier
	infected  int          // how many people have been infected at the location
	visited   int          // how many people have visited the location
}

// Channels for sending statistics between different modules and the statisticsHandler
type statisticsChannels struct {
	parameters chan string
	infection  chan logInfo
	location   chan logInfo
	person     chan logInfo
	done       chan int
	exit       chan bool
}

// Struct holding simulation parameters
type parameters struct {
	numLocations         int
	numPersons           int
	spawnRisk            float64
	spreadRisk           float64
	simLimit             int
	maskUsage            float64
	maskEffectiveness    float64
	vaccineUsage         float64
	vaccineEffectiveness float64
}

// Struct used to sent statistics information
type logInfo struct {
	id   int
	time int
	info string
}

// Return struct for InfectionWorker
type infectionReturn struct {
	persons []Person
	num     int
}

//////////////// Backend Functions ////////////////////

// Goroutine handling the collection of statistics.
// Listens to various channels where it gets reports on what happens during the simulation
// Terminates at the end of the simulation, and is re-launched if another simulation is initiated
func statisticsHandler() {
	//Initial strings
	parameterString := "Simulation parameters:\n"
	infectionString := "Infection Statistics:\n"
	locationString := "Location Statistics:\n"
	personString := "Person Statistics:\n"
	//Maps used to structure result
	infectionMap := make(map[int](map[int]string))
	personMap := make(map[int]string)
	for i := 1; i <= paramStruct.simLimit; i++ {
		tmp := make(map[int]string)
		infectionMap[i] = tmp
	}
	var parameterInfo string
	var personInfo, locationInfo, infectionInfo logInfo
	for {
		select {
		case <-statistics.exit:
			fmt.Println("Exiting statistics handler")
			return
		case length := <-statistics.done: // Indicates end of simulation and triggers writing stats to log file
            locationMap := make(map[int]string)
			for i := 0; i < length; i++ {
				locationInfo = <-statistics.location
				locationMap[locationInfo.id] = locationInfo.info
			}
            for i := 0; i < length; i++ {
				locationString += locationMap[i]
			}
			fmt.Println("Quit listening for stats")
			fmt.Printf("Vaccine prevented %d infections\n", vaccinePrevention)
			fmt.Printf("Vaccine failed to prevent %d infections\n", vaccineNoPrevention)
			for i := 1; i <= paramStruct.simLimit; i++ {
				tmp := infectionMap[i]
				infectionString += fmt.Sprintf("\tTime %d\n", i)
				for j := 0; j < paramStruct.numLocations; j++ {
					infectionString += tmp[j]
				}
			}
			for i := 0; i < paramStruct.numPersons; i++ {
				personString += personMap[i]
			}
			result := parameterString + infectionString + locationString + personString
			path := "log/log.txt"
			err := ioutil.WriteFile(path, []byte(result), 0644)
			if err != nil {
				log.Fatal(err)
			}
			return
		case parameterInfo = <-statistics.parameters: // Simulation parameters, collected once per simulation
			parameterString += parameterInfo
		case infectionInfo = <-statistics.infection:  // Infection information, collected once per location per time step
			infectionMap[infectionInfo.time][infectionInfo.id] = infectionInfo.info
		case personInfo = <-statistics.person:        // Information on each person, collected once per person at the end of their journey
			personMap[personInfo.id] = personInfo.info
		}
	}
}

// Initializes simulation using the parameters input into the GUI.
func InitiateSimulation(window *gotron.BrowserWindow) {
    // Cleanup step, to make sure the GUI knows the previous simulation is over
	data.SendSimulationDone(window, "simulationDone", "end of simulation")
	// Cleanup step to reset backend variables etc. from last simulation
	cleanUp()
	allLocations = nil
    // Channels used 
    uiInfectedChannel := make(chan []data.UpdatePersonUI)
	uiMovementChannel := make(chan []data.UpdatePersonUI)
	uiUpdateChannels := []chan []data.UpdatePersonUI{uiInfectedChannel, uiMovementChannel}
	var initInfoForUI []data.UpdatePersonUI
	exitCh := make(chan bool)
	statistics.exit = exitCh
    //Start background goroutines that receive statistics and sends info to UI
	go statisticsHandler()
	go uiInformationGatherer(uiUpdateChannels)
	// initiate locations
	for i := 0; i < paramStruct.numLocations; i++ {
		synchChannel := make(chan Synch)
		doneChannel := make(chan bool, 2)
		loc := createLocation(locationNames[i], locationModifier[i])
		synchChannels = append(synchChannels, synchChannel)
		doneChannels = append(doneChannels, doneChannel)
		go locationWorker(loc, synchChannel, doneChannel, uiUpdateChannels)
	}
	//Initiate persons
	startTime := time.Now()
	infectionRandomizer := getRandomizer()
	for i := 0; i < paramStruct.numPersons; i++ {
		p := spawnPerson(infectionRandomizer)
        allLocations[p.nextStop].receiveCh <- p
		info := data.UpdatePersonUI{PersonID: p.personID, CurrentLocation: p.nextStop, Infected: p.infected, Masked: p.masked, Vaccinated: p.vaccinated}
		initInfoForUI = append(initInfoForUI, info)
	}
	elapsedTime := time.Since(startTime)
	fmt.Printf("Creating persons took %s\n", elapsedTime)
	var initLocations data.UiLocation
	initLocations.Num = len(allLocations)
	for _, loc := range allLocations {
		initLocations.Names = append(initLocations.Names, loc.Name)
	}
    // send location data and person data to frontend
	data.SendUILocations(window, initLocations)
	data.SendUIUpdate(window, initInfoForUI)
    // send parameter data to statistics handler
	parameterString()
	m := data.StringMessage{}
    // Wait for UI to get ready
	data.ReceiveAndBlock(window, "timeSynchUI", &m)
    data.SendSimulationState(window, false)
}

// Constructor function for person struct
// Will create a certain number of persons and send them back to initiateSimulation
func spawnPerson(rand *rand.Rand) Person {
	infected := false
	infectedTime := -2     //-2 = never infected
	infectedLocation := -2 //-2 = never infected
	vaccine := false
	mask := false
	if rand.Float64() < paramStruct.maskUsage {
		mask = true
	}
    odds := rand.Float64()
	//checks if person should spawn as infected
	if odds < paramStruct.spawnRisk {
		infected = true
		infectedTime = -1     //-1 = infected at spawn
		infectedLocation = -1 //-1 = infected at spawn
	} else if rand.Float64() < paramStruct.vaccineUsage {
        //since vaccines only protect the person themself, only check
        //for vaccination status if they aren't infected to begin with
		vaccine = true
	}
    // Generate randomized path
	numberOfDestinations := rand.Intn(paramStruct.numLocations) + 3
	if numberOfDestinations > paramStruct.numLocations {
		numberOfDestinations = paramStruct.numLocations
	}
	if numberOfDestinations > paramStruct.simLimit{
		numberOfDestinations = paramStruct.simLimit 
	}
    locs := make([]int, len(allLocations)) //generate slice of all locations
    for i := 0; i < len(allLocations); i++{
        locs[i] = i
    }
    rand.Shuffle(len(locs), func (i, j int) {locs[i], locs[j] = locs[j], locs[i]}) //shuffle slice
    path := locs[:numberOfDestinations] //slice slice to proper length
    // Fetch ID
	idMutex.Lock()
	pID := id
	id += 1
	idMutex.Unlock()
	//create person struct
	p := Person{
		personID:         pID,
		infected:         infected,
		infectedTime:     infectedTime,
		infectedLocation: infectedLocation,
		departureTime:    0,
		path:             path,
		done:             false,
		nextStop:         path[0],
		locationsVisited: 0,
		vaccinated:       vaccine,
		masked:           mask,
    }
    return p
}

// Constructor function for location struct
func createLocation(name string, modifier float64) location {
	receiveChannel := make(chan Person, paramStruct.numPersons)
	loc := location{receiveCh: receiveChannel, Name: name, modifier: modifier}
	allLocations = append(allLocations, loc)
	return loc
}

// Updates departure time and departure destination of the person.
func handlePerson(person Person) Person {
	//This part calculates departure time
	randomizer := getRandomizer()
	rem := len(person.path) - person.locationsVisited
	remTime := paramStruct.simLimit - person.departureTime
	maxWait := remTime / rem
	if maxWait == 0 {
		maxWait = 1
	}
	departureTime := randomizer.Intn(maxWait) + person.departureTime + 1
	if departureTime > paramStruct.simLimit {
		departureTime = paramStruct.simLimit //cap it so that a person doesn't stay longer than simulation limit
	}
	person.departureTimes = append(person.departureTimes, departureTime)

	//Update the ticker of visited locations
	person.locationsVisited++
	//check if person is done
	if person.locationsVisited == len(person.path) {
		person.done = true
		person.departureTime = departureTime
		return person
	}

	//Here we choose the next destination in the person's path
	destination := person.path[person.locationsVisited]
	//Update the Person struct
	person.departureTime = departureTime
	person.nextStop = destination
	return person
}

// Function handling the logic for managing a location
// Receives visitors constantly, and starts goroutines that will deal with infection and movement once per time tick.
func locationWorker(locInfo location, synchChannel <-chan Synch, doneChannel <-chan bool, uiChannels []chan []data.UpdatePersonUI) {
	var currentVisitors []Person
	var newVisitors []Person
	receiveChannel := locInfo.receiveCh
	infectionReportCh := make(chan infectionReturn)
	movementReportCh := make(chan []Person)
	var synch Synch

	for {
		select {
		case synch = <-synchChannel: //läser in timeticks
			currentVisitors = append(currentVisitors, newVisitors...) //combine visitor slices
			newVisitors = nil                                         //reset the newVisitors slice
			go infectionWorker(currentVisitors, synch.time, infectionReportCh, uiChannels[0], locInfo.Name, locInfo.modifier)
		case infectionReport := <-infectionReportCh:
			currentVisitors = infectionReport.persons
			locInfo.infected += infectionReport.num
			go locationSender(currentVisitors, locInfo.Name, synch.time, movementReportCh, uiChannels[1])
		case currentVisitors = <-movementReportCh:
			synch.waitGroup.Done()
		case visitor := <-receiveChannel:
			visitor = handlePerson(visitor)
			locInfo.visited++ // increment counter for people who visits the location
			newVisitors = append(newVisitors, visitor)
		case <-doneChannel:
			fmt.Printf("Loc %s received done from EventLoop\n", locInfo.Name)
			locationString(locInfo.Name, locInfo.infected, locInfo.visited)
			return
		default:
			time.Sleep(1 * time.Millisecond)

		}
	}
}

//Sends away persons in persons to their next location if their departureTime == syncTime
//sends back a slice with all the persons who were not sent away to the next location
func locationSender(persons []Person, locationName string, time int, report chan []Person, uiChannel chan []data.UpdatePersonUI) {
	var personsRemaining []Person
	var personsLeaving []data.UpdatePersonUI
	for _, p := range persons {
		if p.departureTime == time {
			if p.done || time == paramStruct.simLimit {
				ui := data.UpdatePersonUI{
					PersonID:        p.personID,
					CurrentLocation: -2, // Person should leave the simulation (go home)
					Infected:        p.infected,
					Masked:          p.masked,
					Vaccinated:      p.vaccinated,
				}
				personString(p, time, locationName)
				personsLeaving = append(personsLeaving, ui)
			} else {
				allLocations[p.nextStop].receiveCh <- p
				ui := data.UpdatePersonUI{
					PersonID:        p.personID,
					CurrentLocation: p.nextStop,
					Infected:        p.infected,
					Masked:          p.masked,
					Vaccinated:      p.vaccinated,
				}
				personsLeaving = append(personsLeaving, ui)
			}
		} else {
			personsRemaining = append(personsRemaining, p)
		}
	}
	report <- personsRemaining
	uiChannel <- personsLeaving
}

// Calculates which persons should be infected in a given sample of persons
func infectionWorker(
	visitors []Person, //the sample of persons
	time int, //the current time
	reportChannel chan infectionReturn, //where to send results
	uiChannel chan []data.UpdatePersonUI, //where to send information GUI needs
	locName string, //which location is the parent for this goroutine
	locmod float64) { //location modifier
	randomizer := getRandomizer()
	var infected []Person
	var healthy []Person
	var changed []data.UpdatePersonUI //slice of those who had their infection status changed this cycle and needs to be updated in GUI
	var result []Person
	var infectedCounter int
	var maskCounter int
	//Checks how many infected are in the sample, and how many of the infected has masks on
	for _, p := range visitors {
		if p.infected {
			if p.masked {
				maskCounter++
			}
			infected = append(infected, p)
		} else {
			healthy = append(healthy, p)
		}
	}
	//calculates density of infected persons to total, adjusted by mask usage/effectiveness
	densSick := (float64(maskCounter)*(1-paramStruct.maskEffectiveness) + float64(len(infected)-maskCounter)) / float64(len(visitors))
	//calculates density of healthy persons to total
	densHealthy := float64(len(healthy)) / float64(len(visitors))
	interferensSickHealthy := densSick * densHealthy
	//adjusts by location modifier and spreadrisk parameter
	risk := interferensSickHealthy * paramStruct.spreadRisk * locmod
	// Goes through the healthy persons and checks if they become infected
	for _, p := range healthy {
		roll := randomizer.Float64()
		if roll < risk { //check if person should get infected
			if p.vaccinated { //check if vaccine prevents infection
				if roll >= risk*(1-paramStruct.vaccineEffectiveness) {
					result = append(result, p)
					vaccinePrevMutex.Lock()
					vaccinePrevention++
					vaccinePrevMutex.Unlock()
					continue
				} else {
					vaccineNoPrevMutex.Lock()
					vaccineNoPrevention++
					vaccineNoPrevMutex.Unlock()
				}
			}
			p.infected = true
			p.infectedTime = time
			p.infectedLocation = getLocIndex(locName)
			ui := data.UpdatePersonUI{
                PersonID:        p.personID,
				CurrentLocation: -1, //-1 = no location change
				Infected:        true,
				Masked:          p.masked,
				Vaccinated:      p.vaccinated,
			}
			changed = append(changed, ui)
			infectedCounter++
		}
		result = append(result, p)
	}
	result = append(result, infected...)
	reportChannel <- infectionReturn{persons: result, num: infectedCounter} //report back to LocationWorker
	uiChannel <- changed //send GUI data
	infectionString(changed, time, locName) //create string for statistics
}

// Reads settings packet from UI and sets the simulation parameters accordingly
func ReadSettings(msg *data.StringMessage) {
	if msg.EventType == "sendSettings" {
		var err error
		paramStruct.numLocations, err = strconv.Atoi(msg.EventValue[0])
		if err != nil {
			fmt.Println("Error: Invalid type for settings value numLocations")
			os.Exit(1)
		} else if paramStruct.numLocations > len(locationNames) {
			paramStruct.numLocations = len(locationNames)
		}
		paramStruct.numPersons, err = strconv.Atoi(msg.EventValue[1])
		if err != nil {
			fmt.Println("Error: Invalid type for settings value numPersons")
			os.Exit(1)
		}
		paramStruct.spawnRisk, err = strconv.ParseFloat(msg.EventValue[2], 64)
		if err != nil {
			fmt.Println("Error: Invalid type for settings value spawnRisk")
			os.Exit(1)
		}
		paramStruct.spreadRisk, err = strconv.ParseFloat(msg.EventValue[3], 64)
		if err != nil {
			fmt.Println("Error: Invalid type for settings value spreadRisk")
		}
		paramStruct.simLimit, err = strconv.Atoi(msg.EventValue[4])
		if err != nil {
			fmt.Println("Error: Invalid type for settings value simLimit")
			os.Exit(1)
		}
		paramStruct.maskUsage, err = strconv.ParseFloat(msg.EventValue[5], 64)
		if err != nil {
			fmt.Println("Error: Invalid type for settings value maskUsage")
			os.Exit(1)
		}
		paramStruct.maskEffectiveness, err = strconv.ParseFloat(msg.EventValue[6], 64)
		if err != nil {
			fmt.Println("Error: Invalid type for settings value maskEffectiveness")
			os.Exit(1)
		}
		paramStruct.vaccineUsage, err = strconv.ParseFloat(msg.EventValue[7], 64)
		if err != nil {
			fmt.Println("Error: Invalid type for settings value vaccineUsage")
			os.Exit(1)
		}
		paramStruct.vaccineEffectiveness, err = strconv.ParseFloat(msg.EventValue[8], 64)
		if err != nil {
			fmt.Println("Error: Invalid type for settings value vaccineEffectiveness")
			os.Exit(1)
		}
		fmt.Printf("Number of locations: %d \nNumber of persons: %d \nSpawn infection risk: %.3f \nSpread Infection risk: %.3f \nSimulation tick limit: %d\nMask Usage: %.3f \nMask effectiveness: %.3f \nVaccine Usage: %.3f \nVaccine Effectiveness: %.3f \n",
			paramStruct.numLocations, paramStruct.numPersons, paramStruct.spawnRisk, paramStruct.spreadRisk, paramStruct.simLimit, paramStruct.maskUsage, paramStruct.maskEffectiveness, paramStruct.vaccineUsage, paramStruct.vaccineEffectiveness)
	} else {
		fmt.Println("Error: not settings packet")
		os.Exit(1)
	}
}

// The event loop and main synchronization enforcer of the simulation. 
func EventLoop(window *gotron.BrowserWindow, resetCh chan bool) {
	// logic for resetting simulation mid-run
	data.SimActive = true
	exitLoop := false
	go func() {
		<-resetCh
		exitLoop = true
		data.SimActive = false
		statistics.exit <- true
	}()
	//time tick for the global time
	timeTick := 1
	for {
		//check if simulation should stop
		if exitLoop {
			cleanUp()
			return
		}
		if timeTick > paramStruct.simLimit {
			time.Sleep(1 * time.Second)
			statistics.done <- len(allLocations)
			cleanUp()
			return
		}
		fmt.Println("========== Current time is", timeTick, "===========")
		//Synchronization portion
		var wg sync.WaitGroup // Instantiate wait group for this tick
		synchInfo := Synch{time: timeTick, waitGroup: &wg}
		wg.Add(len(allLocations)) //Increment wg to appropriate number
		// Send wg+time to all locations
		for i := range synchChannels {
			synchChannels[i] <- synchInfo
		}
		wg.Wait() // wait for completion
		updateInfo := <-transmitToUI //receive UI info
		data.SendUIUpdate(window, updateInfo) //send UI info
		m := data.StringMessage{} 
		data.ReceiveAndBlock(window, "timeSynchUI", &m) // receive ready signal from UI
		timeTick++ //increment time for next loop
		time.Sleep(300 * time.Millisecond) //sleep to smoothen out simulation (for visual purposes)
	}
}

/////HELPER FUNCTIONS/////

// Performs necessary clean-up to run a new simulation. This includes resetting 
// global variables/channels and closing all locationworkers.
func cleanUp() {
	id = 0
	fmt.Println(len(doneChannels))
	for i := range doneChannels {
		fmt.Println("Sending done to location", i)
		doneChannels[i] <- true
		close(doneChannels[i])
		close(synchChannels[i])
	}
	synchChannels = nil
	doneChannels = nil
}

// checks if a UpdatePersonUI-slice contains a certain UpdatePersonUI-object.
func containsBoolForUI(slice []data.UpdatePersonUI, id int) bool {
	for _, a := range slice {
		if a.PersonID == id {
			return true
		}
	}
	return false
}

// Returns a randomizer used to generate PRNG values
func getRandomizer() *rand.Rand {
    time := time.Now().UnixNano()
    source := rand.NewSource(time)
	randomizer := rand.New(source)
	return randomizer
}

// From a location name, gets the index of the location
func getLocIndex(name string) int {
	for i, loc := range allLocations {
		if loc.Name == name {
			return i
		}
	}
	return -99
}

////END OF HELPER FUNCTIONS////

////UI COMMUNICATION FUNCTIONS////

// This function gathers all updated statuses for persons (i.e., which persons have been infected, and which persons have changed location),
// and transmits that information to the frontend. Runs once per time tick.
func uiInformationGatherer(uiChannels []chan []data.UpdatePersonUI) {
	infectionChannel := uiChannels[0]
	movementChannel := uiChannels[1]
	var moved []data.UpdatePersonUI
	var infected []data.UpdatePersonUI
	var result []data.UpdatePersonUI

	for {
		for i := 0; i < paramStruct.numLocations; i++ {
			m := <-infectionChannel
			infected = append(infected, m...)
		}
		for i := 0; i < paramStruct.numLocations; i++ {
			m := <-movementChannel
			moved = append(moved, m...)
		}
		for _, p := range moved {
			if containsBoolForUI(infected, p.PersonID) {
				p.Infected = true
				result = append(result, p)
			} else {
				result = append(result, p)
			}
		}
		for _, p := range infected {
			if containsBoolForUI(result, p.PersonID) {
				continue
			} else {
				result = append(result, p)
			}
		}
		transmitToUI <- result
		moved = nil
		infected = nil
		result = nil
	}
}

///////STATISTICS FUNCTIONS//////

// Function that creates the statistics string of the infection changes for a given time tick
func infectionString(changed []data.UpdatePersonUI, time int, locationName string) {
	locIndex := getLocIndex(locationName)
	if locIndex == -99 {
		fmt.Println("Couldn't fetch location index; no such location name was found")
		return
	}
	if len(changed) == 0 {
		statistics.infection <- logInfo{id: locIndex, time: time, info: fmt.Sprintf(
			"\t\tLocation #%d (%s): No persons were infected.\n",
			locIndex+1,
			locationName,
		)}
		return
	}
	result := fmt.Sprintf(
		"\t\tLocation #%d (%s), the following persons were infected:\n",
		locIndex+1,
		locationName,
	)
	for _, p := range changed {
		str := fmt.Sprintf("\t\t\t#%d\n", p.PersonID)
		result += str
	}
	statistics.infection <- logInfo{id: locIndex, time: time, info: result}
}

// Function that creates the statistics string for a given location
func locationString(locName string, infected int, visited int) {
	index := getLocIndex(locName)
	result := fmt.Sprintf("\tLocation #%d (%s):\n", index + 1, locName)
	// Number of people visited
	result += fmt.Sprintf("\t\tTotal persons visited: %d\n", visited)
	// Number of people infected
	result += fmt.Sprintf("\t\tTotal persons infected: %d\n", infected)
	statistics.location <- logInfo{id: index, time: -1, info: result}
}

// Function that creates the statistics string of simulation parameters
func parameterString() {
	numLoc := "\tNumber of locations: " + strconv.Itoa(paramStruct.numLocations) + "\n"
	numPers := "\tNumber of persons: " + strconv.Itoa(paramStruct.numPersons) + "\n"
	spawnRiskNr := fmt.Sprintf("%.3f", paramStruct.spawnRisk)
	spreadRiskNr := fmt.Sprintf("%.3f", paramStruct.spreadRisk)
	maskUsageNr := fmt.Sprintf("%.3f", paramStruct.maskUsage)
	maskEffectiveNr := fmt.Sprintf("%.3f", paramStruct.maskEffectiveness)
	vaccineUsageNr := fmt.Sprintf("%.3f", paramStruct.vaccineUsage)
	vaccineEffectiveNr := fmt.Sprintf("%.3f", paramStruct.vaccineEffectiveness)
	spawnRisk := "\tInfection risk on spawn: " + spawnRiskNr + "\n"
	spreadRisk := "\tInfection spread risk: " + spreadRiskNr + "\n"
	simLimit := "\tSimulation time limit: " + strconv.Itoa(paramStruct.simLimit) + "\n"
	maskUsage := "\tMask usage: " + maskUsageNr + "\n"
	maskEffective := "\tMask effectiveness: " + maskEffectiveNr + "\n"
	vaccineUsage := "\tVaccine usage: " + vaccineUsageNr + "\n"
	vaccineEffective := "\tVaccine effectiveness: " + vaccineEffectiveNr + "\n"
	result := numLoc + numPers + spawnRisk + spreadRisk + simLimit + maskUsage + maskEffective + vaccineUsage + vaccineEffective
	statistics.parameters <- result
}

// Function that creates statistics string for the given person
func personString(p Person, time int, locName string) {
	result := fmt.Sprintf("\tPerson #%d:\n", p.personID+1)
	var infected string
	if p.infected {
		if p.infectedLocation == -1 {
			infected = "\t\tWas infected on spawn.\n"
		} else {
			infected = fmt.Sprintf("\t\tWas infected at time %d at location #%d (%s).\n",
				p.infectedTime,
				p.infectedLocation+1,
				locationNames[p.infectedLocation])
		}
	} else {
		infected = "\t\tWas not infected.\n"
	}
	result += infected
	var path string
	for i, loc := range p.path {
		if i == 0 {
			path += fmt.Sprintf("\t\tThey started at location #%d (%s),\n\t\tthen moved to ",
				loc+1, allLocations[loc].Name)
		} else {
			path += fmt.Sprintf("location #%d (%s) at time %d", loc+1, allLocations[loc].Name, p.departureTimes[i-1])
			if i+1 == len(p.path) {
				path += ",\n\t\twhich was their final destination.\n"
			} else {
				path += ",\n\t\tthen to "
			}
		}
	}
	result += path
	statistics.person <- logInfo{id: p.personID, info: result}
}
