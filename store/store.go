package store

import (
	"sync"
)

var (
	scenarios = []Scenario{
		{ID: 1, Name: "API Stress Test", Description: "Тестирование API", Tags: []string{"API", "Stress"}},
		{ID: 2, Name: "Login Test", Description: "Тест авторизации", Tags: []string{"Auth"}},
	}
	loadProfiles = []LoadProfile{
		{ID: 1, Name: "High Load", Type: "Constant"},
		{ID: 2, Name: "Ramp-Up", Type: "Incremental"},
	}
	testRuns = []TestRun{
		{ID: 1, Name: "API Stress Run 1", StartTime: "2025-05-18 10:00", Status: "Completed"},
	}
	mutex sync.Mutex
)

func ScenarioCount() int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(scenarios)
}

func GetScenario(index int) Scenario {
	mutex.Lock()
	defer mutex.Unlock()
	return scenarios[index]
}

func AddScenario(scenario Scenario) {
	mutex.Lock()
	defer mutex.Unlock()
	scenarios = append(scenarios, scenario)
}

func LoadProfileCount() int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(loadProfiles)
}

func GetLoadProfile(index int) LoadProfile {
	mutex.Lock()
	defer mutex.Unlock()
	return loadProfiles[index]
}

func AddLoadProfile(profile LoadProfile) {
	mutex.Lock()
	defer mutex.Unlock()
	loadProfiles = append(loadProfiles, profile)
}

func TestRunCount() int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(testRuns)
}

func GetTestRun(index int) TestRun {
	mutex.Lock()
	defer mutex.Unlock()
	return testRuns[index]
}

func AddTestRun(testRun TestRun) {
	mutex.Lock()
	defer mutex.Unlock()
	testRuns = append(testRuns, testRun)
}
