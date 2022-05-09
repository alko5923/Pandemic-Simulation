package backend

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ospp-projects-2021/clockwork/src/data"
)

func cleanUpTest() {
	id = 0
	allLocations = nil
	paramStruct.numLocations = 1
}

//Tests for func SpawnPerson(odds float64) Person
//Tests the correctness of the IDs of 3 spawned persons
func TestSpawnPersonId(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestSpawnPersonId~~~~~")
	paramStruct.numLocations = 3
	var tests = []struct {
		want int
	}{
		{0},
		{1},
		{2},
	}
	createLocation("ICA", 0)
	createLocation("ICA", 0)
	createLocation("ICA", 0)
	//paramStruct.numLocations = 3
	paramStruct.simLimit = 1
	source := rand.NewSource(1)
	randomizer := rand.New(source) //1 = 0.6 Not infected // 2 = 0.1 Infected
	paramStruct.spawnRisk = 0.5
	for _, tt := range tests {

		testname := "test"
		t.Run(testname, func(t *testing.T) {
			ans := spawnPerson(randomizer)

			ansID := ans.personID

			if ansID != tt.want {
				t.Errorf("got %d, want %d", ansID, tt.want)
			}
		})
	}
}

//Test for func CreateLocation(name string) Location
//Tests the correctness of names in locations
func TestCreateLocation(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestCreateLocation~~~~~")
	var tests = []struct {
		name         string
		modifier     float64
		wantName     string
		wantModifier float64
	}{
		{"ICA", 0.1, "ICA", 0.1},
		{"Apoteket", 1.4, "Apoteket", 1.4},
		{"Zoo", 1, "Zoo", 1},
	}

	for _, tt := range tests {
		testname := tt.name
		t.Run(testname, func(t *testing.T) {
			ans := createLocation(tt.name, tt.modifier)
			if ans.Name != tt.wantName {
				t.Errorf("LocationName TEST FAIL: got %s, want %s", ans.Name, tt.wantName)
			} else if ans.modifier != tt.wantModifier {
				t.Errorf("Modifier TEST FAIL: got %f, want %f", ans.modifier, tt.wantModifier)
			}
		})
	}
}

//Remove function used by DuplicateExists
func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

//Checks if a slice of ints contains duplicate elements
func DuplicatesExists(path []int) bool {
	defer cleanUpTest()
	for i, location := range path {
		for _, locationSearch := range remove(path, i) {
			if location == locationSearch {
				return true
			}
		}
	}
	return false
}

//Test for func SpawnPerson(odds float64) person
//Tests the default values that a newly created person should have
func TestSpawnPersonInit(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestSpawnPersonInit~~~~~")
	path1 := []int{2, 4, 6}
	source := rand.NewSource(2)
	randomizer := rand.New(source)
	NumberOfLocations := 3
	var tests = []struct {
		odds             *rand.Rand
		infectedTime     int
		path             []int
		done             bool
		NextStop         int
		LocationsVisited int
	}{
		{randomizer, -1, path1, false, 1, 0},
	}
	paramStruct.numLocations = 1
	createLocation("Ica", 0)
	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.odds)
		t.Run(testname, func(t *testing.T) {
			person := spawnPerson(tt.odds)
			ansInfectedTime := person.infectedTime
			ansPath := person.path
			ansDone := person.done
			ansLocationsVisited := person.locationsVisited
			if ansInfectedTime != tt.infectedTime {
				t.Errorf("got %d, want %d", ansInfectedTime, tt.infectedTime)
			}
			var lengthAnsPath = len(ansPath)
			if lengthAnsPath > NumberOfLocations {
				t.Errorf("Too many locations: Got %d, Limit; %d", lengthAnsPath, NumberOfLocations)
				if DuplicatesExists(ansPath) {
					t.Errorf("Duplicates Exists in path")
				}
			}
			if ansDone != tt.done {
				t.Errorf("got %v, want %v", ansDone, tt.done)
			}
			if ansLocationsVisited != tt.LocationsVisited {
				t.Errorf("got %d, want %d", ansLocationsVisited, tt.LocationsVisited)
			}
		})
	}
}

// Helper function for TestLocationWorker
// Mimicks location2 and waits for person1 to be sent in the channel
func testLocationWorkerHelpFunction(loc2 location, personChannel chan Person) {
	information := <-loc2.receiveCh
	personChannel <- information

}

// Test for func LocationWorker(locInfo Location, syncChannel <-chan Synch, uiChannels []chan []data.UpdatePersonUI)
// Tests that locationWorker uses locationSender propably and sends one person to next location
func TestLocationWorker(t *testing.T) {
	fmt.Println("~~~~~TestLocationWorker~~~~~")
	defer cleanUpTest()
	paramStruct.simLimit = 10
	Time := 1
	var synch Synch
	var wg sync.WaitGroup
	synch.time = Time
	synch.waitGroup = &wg
	syncChannel := make(chan Synch)
	loc1 := createLocation("ICA", 0)
	loc2 := createLocation("Chipsfabriken", 0)
	source := rand.NewSource(2)
	randomizer := rand.New(source)
	person1 := spawnPerson(randomizer)
	person1.departureTime = 1
	person1.path = nil
	person1.path = append(person1.path, 0)
	person1.path = append(person1.path, 1)
	person1.departureTime = 5
	maxDeparture := 10 // The max value of departureTime, what is it really??

	var uiChannels []chan []data.UpdatePersonUI
	uiChannels = append(uiChannels, make(chan []data.UpdatePersonUI))
	uiChannels = append(uiChannels, make(chan []data.UpdatePersonUI))
	personChannel := make(chan Person)
	go testLocationWorkerHelpFunction(loc2, personChannel)
	doneChannel := make(chan bool, 2)
	doneChannels = append(doneChannels, doneChannel)
	go locationWorker(loc1, syncChannel, doneChannel, uiChannels)

	loc1.receiveCh <- person1
	time.Sleep(50 * time.Millisecond) // This is to make sure that person1 doesn't get a departure time thats lower than the time tick

	for i := 0; i < maxDeparture; i++ {
		wg.Add(1)
		synchInfo := Synch{time: Time, waitGroup: &wg}
		syncChannel <- synchInfo

		Time += 1
	}

	select {
	case personInfo := <-personChannel:
		fmt.Sprintf("%d", personInfo.departureTime) // this is to make the compiler happy by using personInfo atleast once :)

	default:
		t.Errorf("One person should have left during this time but didn't, NOTE: This could fail, if it keeps on failing after multiple runs it could be something wrong\n")
	}

}

// Test for func LocationSender(persons []Person, synchInfo Synch, report chan []Person, uiChannel chan []data.UpdatePersonUI)
// Tests that a person leaves at timetick 3
func TestLocationSenderOnePerson(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestLocationSenderOnePerson~~~~~")
	paramStruct.simLimit = 3
	//create synch
	time := 1
	var synch Synch
	var wg sync.WaitGroup
	synch.time = time
	synch.waitGroup = &wg
	//create channels
	var remained chan []Person = make(chan []Person)
	var left chan []data.UpdatePersonUI = make(chan []data.UpdatePersonUI)
	//specify how many locations we want for this test
	paramStruct.numLocations = 1
	//create the location
	loc1 := createLocation("Ica", 0)
	//initiate person
	var persons []Person
	source := rand.NewSource(2)
	randomizer := rand.New(source)

	p := spawnPerson(randomizer)
	//set a specific departure time
	p.departureTime = 3
	persons = append(persons, p)

	//loop 3 times, check that person leaves at time = 3
	for i := 0; i < 3; i++ {

		wg.Add(1)

		go locationSender(persons, loc1.Name, synch.time, remained, left)

		rem := <-remained

		if i == 2 {
			if len(rem) != 0 { //check that person left properly
				t.Errorf("Person remained when it shouldn't have (time %d)\n", time)
			}
		} else {
			if len(rem) != 1 { //check that person didn't leave when it shouldn't
				t.Errorf("Person did not remain as it should (time %d)\n", time)
			}

		}
		gone := <-left
		if i == 2 {
			if len(gone) != 1 { //check that person is in outgoing slice as it should
				t.Errorf("Person is not in outgoing slice when it should be (time %d)\n", time)
			} else {
				if gone[0].PersonID != p.personID {
					t.Errorf("Person in gone slice does not match person sent to Sender\n")
				}
			}
		} else {
			if len(gone) != 0 {
				t.Errorf("Person left when it shouldn't (time %d)\n", time)
			}
		}
		time++
		synch.time = time
	}
}

// Test for func LocationSender(persons []Person, synchInfo Synch, report chan []Person, uiChannel chan []data.UpdatePersonUI)
// Tests that one person leaves at each time tick, and it is the correct person that leaves
func TestLocationSenderTenPersons(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestLocationSenderTenPersons~~~~~")
	paramStruct.simLimit = 20
	paramStruct.numPersons = 10
	//create synch
	time := 1
	var synch Synch
	var wg sync.WaitGroup
	synch.time = time
	synch.waitGroup = &wg
	//create channels
	var remained chan []Person = make(chan []Person)
	var left chan []data.UpdatePersonUI = make(chan []data.UpdatePersonUI)
	//specify how many locations we want for this test
	paramStruct.numLocations = 2
	//create the location
	loc1 := createLocation("Ica", 0)
	createLocation("Willys", 0)
	source := rand.NewSource(1)
	randomizer := rand.New(source)

	//initiate persons and set departure time
	var persons []Person
	for i := 0; i < 10; i++ {
		p := spawnPerson(randomizer)
		p.departureTime = i + 1
		persons = append(persons, p)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go locationSender(persons, loc1.Name, synch.time, remained, left)

		persons = <-remained
		if len(persons) != 10-time { //check that person left properly
			t.Errorf("Person remained when it shouldn't have (time %d)\n", time)
		}

		gone := <-left
		if len(gone) != 1 { //check that person is in outgoing slice as it should
			if gone[0].PersonID != time-1 {
				t.Errorf("The PersonID doesn't match with the expected one Got:_%d, Want: %d", gone[0].PersonID, time-1)
			}
			t.Errorf("Person is not in outgoing slice when it should be (time %d)\n", time)
		}

		time++
		synch.time = time
	}
}

// Test for func InfectionWorker (visitors []Person, time int, reportChannel chan<- []Person, uiChannel chan []data.UpdatePersonUI, locName string, locSize int)
// Checks that all persons except person 2,3,4 remain infected. 2,3,4 will always be infected with this hardcoded seed.
func TestInfectionWorkerAllHealthy(t *testing.T) {
	fmt.Println("~~~~TestInfectionWorkerAllHealthy~~~~")
	defer cleanUpTest()
	infectionReportCh := make(chan infectionReturn)
	paramStruct.numLocations = 1
	source := rand.NewSource(1)
	randomizer := rand.New(source)
	loc := createLocation("Healthy", 0)
	var persons []Person
	for i := 0; i < 10; i++ {
		p := spawnPerson(randomizer)
		persons = append(persons, p)
	}
	changed := make(chan []data.UpdatePersonUI)
	go infectionWorker(persons, 1, infectionReportCh, changed, loc.Name, loc.modifier)

	ppl := <-infectionReportCh
	ppl2 := ppl.persons
	if len(ppl2) != len(persons) {
		t.Errorf("Length of return slice does not match number of persons inserted\n")
	}

	for _, p := range ppl2 {
		if p.infected {
			if p.personID != 2 && p.personID != 3 && p.personID != 4 { //These persons will always be infected because of how the infection algorithm works
				t.Errorf("Person #%d should be healthy, but is infected :D\n", p.personID)
			}

		}
	}
	change := <-changed
	if len(change) != 0 {
		t.Errorf("Changed slice should be empty, but it isn't\n")
	}
}

// Test for func InfectionWorker (visitors []Person, time int, reportChannel chan<- []Person, uiChannel chan []data.UpdatePersonUI, locName string, locSize int)
// Checks that all persons remain infected
func TestInfectionWorkerAllInfected(t *testing.T) {
	fmt.Println("~~~~TestInfectionWorkerAllInfected~~~~")
	defer cleanUpTest()
	infectionReportCh := make(chan infectionReturn)
	source := rand.NewSource(1)
	randomizer := rand.New(source)
	paramStruct.numLocations = 1
	loc := createLocation("Infected", 0)
	var persons []Person
	for i := 0; i < 10; i++ {
		p := spawnPerson(randomizer)
		persons = append(persons, p)
	}
	changed := make(chan []data.UpdatePersonUI)
	go infectionWorker(persons, 1, infectionReportCh, changed, loc.Name, loc.modifier)
	ppl := <-infectionReportCh
	ppl2 := ppl.persons
	if len(ppl2) != len(persons) {
		t.Errorf("Length of return slice does not match number of persons inserted\n")
	}
	for _, p := range ppl2 {
		if !p.infected {
			if p.personID != 0 && p.personID != 1 && p.personID != 5 && p.personID != 6 && p.personID != 7 && p.personID != 8 && p.personID != 9 {
				t.Errorf("Persons %d should be infected, but is healthy >:(\n", p.personID)
			}
		}
	}
	change := <-changed
	if len(change) != 0 {
		t.Errorf("Changed slice should be empty, but isn't\n")
	}
}

// Test for func InfectionWorker (visitors []Person, time int, reportChannel chan<- []Person, uiChannel chan []data.UpdatePersonUI, locName string, locSize int)
// Checks that a proper amount of persons become infected, within a certain margin of error (delta)
func TestInfectionWorkerFiftyFifty(t *testing.T) {
	//this function will need to be updated when/if we make a more complex infectionRisk calculator
	fmt.Println("~~~~TestInfectionWorkerFiftyFifty~~~~")
	defer cleanUpTest()
	infectionReportCh := make(chan infectionReturn)
	paramStruct.numLocations = 1
	loc := createLocation("FiftyFifty", 1)
	howManyPeople := 1000
	howManyLoops := 10
	total := 0
	for i := 0; i < howManyLoops; i++ {
		persons := InitPersonsForFiftyFifty(howManyPeople, t)
		//result := make(chan []Person)
		changed := make(chan []data.UpdatePersonUI)
		go infectionWorker(persons, 1, infectionReportCh, changed, loc.Name, loc.modifier) //loc.Size drastically changes the number of people that will be infected
		ppl := <-infectionReportCh
		ppl2 := ppl.persons
		if len(ppl2) != len(persons) {
			t.Errorf("Length of return slice does not match number of persons inserted\n")
		}
		change := <-changed
		//check that length of change slice is equal to delta between incoming and outgoing number of infected
		infected := 0
		for _, p := range ppl2 {
			if p.infected {
				infected++
			}
		}
		delta := infected - howManyPeople/2
		if len(change) != delta {
			t.Errorf("Length of change slice does not match number of people that should have had their status changed")
		}
		total += delta
	}
	average := total / howManyLoops
	expected := 0 //This is the expected number of infections with the current infection algorithm and settings
	delta := math.Abs(float64(average - expected))
	fmt.Printf("Average number of new infected per run is %d.\nExpected value is %d.\nDelta is %.0f\n", average, expected, delta)
	if delta > 10 {
		t.Errorf("Delta is greater than 10. This could be a statistical anomaly, re-run test to see if it repeats.")
	}
}

// Helper function for the tests
// Creates one slice with half of the persons healthy and half infected
func InitPersonsForFiftyFifty(howMany int, t *testing.T) []Person {
	var persons []Person
	//create persons, half infected half not
	source := rand.NewSource(1)
	randomizer := rand.New(source)
	source2 := rand.NewSource(2)
	randomizer2 := rand.New(source2)
	p := spawnPerson(getRandomizer())
	for i := 0; i < howMany; i++ {
		if i%2 == 0 {
			p = spawnPerson(randomizer) //i=0 will be infected, i=1 will be healthy, etc.
			p.infected = true
		} else {
			p = spawnPerson(randomizer2) //i=0 will be infected, i=1 will be healthy, etc.
			p.infected = false
		}

		if i%2 == 0 {
			if !p.infected {
				t.Errorf("Person %d should have spawned as infected\n", i)
			}
		} else {
			if p.infected {
				t.Errorf("Person %d should have spawned as healthy\n", i)
			}
		}
		persons = append(persons, p)
	}
	infected := 0
	healthy := 0
	for _, p := range persons {
		if p.infected {
			infected++
		} else {
			healthy++
		}
	}
	if infected != len(persons)/2 || healthy != len(persons)/2 {
		t.Errorf("Slice should be half infected, half not. Actual distribution:\nInfected: %d, Healthy: %d\n", infected, healthy)
	}
	return persons
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////

//Test for func containsBoolForUI(slice []data.UpdatePersonUI, id int) bool
//Tests if the functions correctly determines if the slice contains a certain PersonId
func TestContainsBoolForUi(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestContainsBoolForUi~~~~~")

	var upui1 data.UpdatePersonUI
	upui1.CurrentLocation = 1
	upui1.Infected = true
	upui1.PersonID = 1

	var upui2 data.UpdatePersonUI
	upui2.CurrentLocation = 2
	upui2.Infected = false
	upui2.PersonID = 2

	slice1 := []data.UpdatePersonUI{upui1, upui2}
	slice2 := []data.UpdatePersonUI{}
	var tests = []struct {
		slice    []data.UpdatePersonUI
		searchId int
		want     bool
	}{
		{slice1, 1, true},
		{slice1, 0, false},
		{slice2, 0, false},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.searchId)
		t.Run(testname, func(t *testing.T) {
			ans := containsBoolForUI(tt.slice, tt.searchId)
			if ans != tt.want {
				t.Errorf("containsBoolForUI didn't return as expected")
			}
		})
	}
}

// Test for func GetRandomizer() *rand.Rand
// Checks that two different calls to it doesn't return the same seed (It is highly unlikely that they will be the same)
func TestGetRandomizer(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestGetRandomnizer~~~~~")

	ans1 := getRandomizer()
	ans2 := getRandomizer()
	if &ans1 == &ans2 {
		t.Errorf("GetRandomizer returned two equivalent seeds")
	}

}

// Test for func getLocIndex(name string) int
// Checks that it returns correct index and -99 if it doesn't exist in allLocations
func TestGetLocIndex(t *testing.T) {
	fmt.Println("~~~~~TestGetLocIndex~~~~~")
	createLocation("Team10House", 0)
	createLocation("MaverickHouse", 0)
	createLocation("Inferno Online", 0)

	ans1 := getLocIndex("Team10House")
	ans2 := getLocIndex("MaverickHouse")
	ans3 := getLocIndex("Inferno Online")
	ans4 := getLocIndex("Rhodos")

	if ans1 != 0 || ans2 != 1 || ans3 != 2 {
		t.Errorf("A location that was expected in allLocations wasn't either found or on wrong index")
	}

	if ans4 != -99 {
		t.Errorf("A location that shouldn't be in allLocations was found in allLocations")
	}

}

// Test for func handlePerson(person Person) Person
// Tests that LocationsVisited and NextStop is being propely updated
func TestHandlePerson(t *testing.T) {
	defer cleanUpTest()
	fmt.Println("~~~~~TestHandlePerson~~~~~")
	source := rand.NewSource(1)
	randomizer := rand.New(source)

	person1 := spawnPerson(randomizer)
	path1 := []int{2}
	person1.path = path1

	person1 = handlePerson(person1)
	if person1.locationsVisited != 1 {
		t.Errorf("Person1's LocationsVisited returned wrong, got: %d, want: %d", person1.locationsVisited, 1)
	}

	person2 := spawnPerson(randomizer)
	path2 := []int{1, 2}
	person2.path = path2
	person2 = handlePerson(person2)
	if person2.nextStop != 2 {
		t.Errorf("Person2's nextStop returned wrong, got: %d, want: %d", person1.nextStop, 2)
	}

	person2 = handlePerson(person2)

	if person2.locationsVisited != 2 {
		t.Errorf("Person2's LocationsVisited returned wrong, got: %d, want: %d", person2.locationsVisited, 2)
	}

}

//Test for func cleanUp()
//Checks that cleanUp resets variables correctly
func TestCleanUp(t *testing.T) {
	fmt.Println("~~~~~TestCleanUp~~~~~")
	defer cleanUpTest()
	doneChannels = nil
	doneChannel := make(chan bool, 1)
	doneChannel2 := make(chan bool, 1)
	doneChannels = append(doneChannels, doneChannel)
	doneChannels = append(doneChannels, doneChannel2)
	synchChannel := make(chan Synch)
	synchChannel2 := make(chan Synch)
	synchChannels = append(synchChannels, synchChannel)
	synchChannels = append(synchChannels, synchChannel2)
	cleanUp()
	for i := range doneChannels {
		ans := <-doneChannels[i]
		if !ans {
			t.Errorf("doneChannel was not set to true")
		}
	}
	if synchChannels != nil {
		t.Errorf("\"synchChannels\" was not set to nil")
	}
	if doneChannels != nil {
		t.Errorf("\"doneChannels\" was not set to nil")
	}
	if id != 0 {
		t.Errorf("id was not set to 0")
	}
}
