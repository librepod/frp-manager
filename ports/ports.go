package ports

import (
	"errors"
	"fmt"
	"main/store"
	"strconv"
	"strings"
	"time"
)

const NO_MORE_PORTS = "NO_MORE_FREE_PORTS_LEFT"

var nextPort func() (int, error)

func init() {
	// TODO: Pass port range from frps.ini config
	allowedPorts := "6008-6010,6017"
	nextPort = createAllowPortsGenerator(allowedPorts)
}

func GetFreePort(userName string) (int, error) {
	fmt.Println("Looking for a free port...")

	userRecord := store.UserRecord{}
	found, dbErr := store.DB.Get(userName, &userRecord)
	if dbErr != nil {
		fmt.Println("error occurred accessing db")
		panic(dbErr)
	}
	if found {
		fmt.Println("Found previously allocated port number: ", userRecord.Port)
		return userRecord.Port, nil
	}

	fmt.Printf("No record in DB for the '%s' user\n", userName)
	fmt.Println("Allocating new port number for the user...")

	freePort := 0
	// Iterate through all the allowedPorts skeeping those that had been already
	// alocated to somebody (have records in DB)
	for p, err := nextPort(); err == nil; p, err = nextPort() {
		fmt.Printf("Trying port %+v...\n", p)
		portRecord := store.PortRecord{}
		found, dbErr := store.DB.Get(strconv.Itoa(p), &portRecord)
		if dbErr != nil {
			fmt.Println("error occurred accessing db")
			panic(dbErr)
		}
		if !found {
			fmt.Println("Found a free port to use: ", p)
			freePort = p
			break
		}
	}

	if freePort == 0 {
		// If we still have zero value port number, this means that we reached our port limits
		return 0, errors.New(NO_MORE_PORTS)
	}

	// Saving the port to DB
	savePortNumber(userName, freePort)

	return freePort, nil
}

// This is a closure that accepts port ranges in string representation like this:
// `3000-8000,60000-65000` and returns an iterator function which returns
// next port number and an error in case if no more ports left from the
// ranges of ports supplied.
func createAllowPortsGenerator(portsRange string) func() (int, error) {
	rangeSlice := strings.Split(portsRange, ",")
	i := 0
	ranges := make([][]int, len(rangeSlice))
	for i, r := range rangeSlice {
		if strings.Contains(r, "-") {
			rangeVals := strings.Split(strings.TrimSpace(r), "-")
			start, _ := strconv.Atoi(rangeVals[0])
			end, _ := strconv.Atoi(rangeVals[1])

			if start > end {
				panic("😱 invalid range supplied")
			}

			ranges[i] = []int{start, end}
		} else {
			port, _ := strconv.Atoi(r)
			ranges[i] = []int{port, port}
		}
	}

	i = ranges[0][0]
	j := 0

	// Closure captures range variables
	return func() (int, error) {
		if i > ranges[j][1] {
			j++
			if j >= len(ranges) {
				j--
				return 0, errors.New(NO_MORE_PORTS)
			}
			i = ranges[j][0]
		}
		val := i
		i++
		return val, nil
	}
}

func savePortNumber(userName string, port int) {
	fmt.Printf("Persisting record to DB: userName=%s, port=%+v...\n", userName, port)

	date := time.Now().UTC()
	ur := store.UserRecord{
		Port:      port,
		CreatedAt: date,
		// IP:        c.ClientIP(),
	}
	pr := store.PortRecord{
		User: userName,
		CreatedAt: date,
		// IP:        c.ClientIP(),
	}

	err := store.DB.Set(userName, ur)
	if err != nil {
		fmt.Printf("Error setting value: %+v.\n", err)
		panic(err)
	}
	err = store.DB.Set(strconv.Itoa(port), pr)
	if err != nil {
		fmt.Printf("Error setting value: %+v.\n", err)
		panic(err)
	}
}
