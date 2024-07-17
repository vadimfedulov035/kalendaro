package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

const Dir string = "/root/kalendaro"
const DirC string = "/root/kalendaro/website/content"
const DirD string = "/root/kalendaro/website/download"
const DirE string = "/root/kalendaro/eternal"

type Gen struct {
    Months [][2]int
    Years []int
}

type Nearest struct {
	Months map[[3]int]map[int][2]int
	Years map[[3]int]map[int]int
}

// checks if year is leap
func isLeap(year int) bool {
	leap4   := year % 4 == 0
	leap100 := year % 100 == 0
	leap400 := year % 400 == 0
	leap := leap4 && (!leap100 || leap400)
	return leap
}

// calculates day from Gregorian date
func calcDay(month int, monthDay int, leap bool) int {
	// set month days
	monthDays := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if leap {
		monthDays[1] += 1
	}
	// add up all days
	day := 0
	for _, mDays := range monthDays[:month-1] {
		day += mDays
	}
	day += monthDay
	return day
}

// calculates IFC date from day
func calcDateI(day int, leap bool) (int, int) {
	const leapDay = 169
	yearDay := 365
	// calculate date
	month := (day / 28) + 1 // 1 2 ... 13
	monthDay := day % 28    // 1 2 ... 0 (<- 28 [will be reassigned])
	// recalculate date according to the leap day
	if leap {
		yearDay += 1
		if day == leapDay {
			// 29.06 (169 day)
			// leap day is irregular
			month = 6
			monthDay = 29
		} else if day > leapDay {
			// 01.07+ (170+ day)
			// all days after the leap day are regular
			afterdays := day - leapDay
			month = 7 + (afterdays / 28)
			monthDay = afterdays % 28
		}
	}
	// recalculate date according to the year day
	if day == yearDay {
		month = 13
		monthDay = 29
	}
	// reassign the start and the end of the cycle
	if monthDay == 0 {
		month--
		monthDay = 28
	}
	return month, monthDay
}

// returns IFC date as Date struct
func getDateI(tzShiftMinutes int) [3]int {
	place := time.FixedZone("UTC", tzShiftMinutes * 60)
	timestamp := time.Now().In(place)
	// get year
	year := timestamp.Year()
	// get Gregorian date
	monthG := int(timestamp.Month())
	dayG := timestamp.Day()
	// calculate IFC date
	leap := isLeap(year)
	day := calcDay(monthG, dayG, leap)
	monthI, dayI := calcDateI(day, leap)
	// unite IFC
	dateI := [3]int{year, monthI, dayI}
	return dateI
}

// wrapper for explicit command execution
func showCmd(app string, args ...string) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(app, args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = "/root/kalendaro", &stdout, &stderr
	// print output of command
	cmd.Run()
	fmt.Println(stdout.String())
	cmd.Wait()
}

// wrapper for silent command execution
func runCmd(app string, args ...string) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(app, args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = "/root/kalendaro", &stdout, &stderr
	// print full command
	fmt.Printf("%s", app)
	for _, arg := range args {
		fmt.Printf(" %s", arg)
	}
	fmt.Printf("\n")
	// try to execute
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Stdout:", stdout.String())
		fmt.Println("Stderr:", stderr.String())
		panic(err)
	}
	cmd.Wait()
}

// set dates according to their timezones
func setTzDates(tzShifts [38]int, tzDates map[int][3]int) {
	for _, tzShift := range tzShifts {
		date := getDateI(tzShift)
		tzDates[tzShift] = date
	}
}

// get unique dates from timezone dates
func getUniqueDays(tzShifts [38]int, tzDates map[int][3]int) [][3]int {
	unique := make(map[[3]int]bool, 3)
	var uDays [][3]int
	for _, tzShift := range tzShifts {
		date := tzDates[tzShift]
		unique[date] = true
	}
	for date, _ := range unique {
		uDays = append(uDays, date)
	}
	fmt.Println(uDays)
    	return uDays
}

// get unique months
func getUniqueMonths(months [][2]int) [][2]int {
	unique := make(map[[2]int]bool)
	var uMonths [][2]int
	for _, month := range months {
		unique[month] = true
	}
	for month, _ := range unique {
		uMonths = append(uMonths, month)
	}
	return uMonths
}

// get unique years
func getUniqueYears(years []int) []int {
	unique := make(map[int]bool)
	var uYears []int
	for _, year := range years {
		unique[year] = true
	}
	for year, _ := range unique {
		uYears = append(uYears, year)
	}
	return uYears
}

// get nearest months and years to unique dates
func getNearest(uDates [][3]int) Nearest {
	// initialize outer maps
	months := make(map[[3]int]map[int][2]int)
	years := make(map[[3]int]map[int]int)
	// initialize inner maps
	for _, date := range uDates {
    		months[date] = make(map[int][2]int)
    		years[date] = make(map[int]int)
	}
	// get previous, current, next months and years for every date
	for _, date := range uDates {
		year, month := date[0], date[1]
		// basic loop
		for i := -1; i < 2; i++ {
			months[date][i] = [2]int{year, month+i}
			years[date][i] = year + i
		}
		// exception case
		if month == 1 {
			months[date][-1] = [2]int{year-1, 13}
		} else if month == 13 {
			months[date][1] = [2]int{year+1, 1}
		}
	}
	nearest := Nearest{months, years}
	return nearest
}

// get all months and years from nearest to generate
func getGen(uDates [][3]int, nearest Nearest) Gen {
	var monthsGen [][2]int
	var yearsGen []int
	for _, date := range uDates {
		for i := -1; i < 2; i++ {
			monthsGen = append(monthsGen, nearest.Months[date][i])
			yearsGen = append(yearsGen, nearest.Years[date][i])
		}
	}
	gen := Gen{Months: monthsGen, Years: yearsGen}
	return gen
}

func getDescription(dates [3]int) string {
	month, day := dates[1], dates[2]
	monthNamesI := [13]string{"januaro", "februaro", "marto",
		"aprilo", "majo", "junio", "sunio", "julio", "aŭgusto",
		"septembro", "oktobro", "novembro", "decembro"}
	descriptionStr := "Hodiaŭ estas la %d de %s"
	description := fmt.Sprintf(descriptionStr, day, monthNamesI[month-1])
	return description
}

// generate new days
func genDays(days [][3]int) {
	tex := fmt.Sprintf("%s/ifc.tex", Dir)
	pdf := fmt.Sprintf("%s/ifc.pdf", Dir)
	for _, day := range days {
		var idRaw string
		for i := 0; i < 3; i++ {
			idRaw += fmt.Sprintf("%02d-", day[i])
		}
		id := idRaw[:len(idRaw)-1]
		f := fmt.Sprintf("/root/kalendaro/.tmp/ifc-%s.png", id)
		if _, err := os.Stat(f); err == nil {
			continue
		}
		runCmd("sed", "-i", "22s/{.*}/{0}/", tex)
		for i := 0; i < 3; i++ {
			line := i + 23
			sedExp := fmt.Sprintf("%ds/{.*}/{%d}/", line, day[i])
			runCmd("sed", "-i", sedExp, tex)
		}
		// calendar
		runCmd("xelatex", tex)
		cmd := exec.Command("pdftoppm", "-png", pdf)
		output, _ := cmd.Output()
		calName := fmt.Sprintf("%s/.tmp/ifc-%s.png", Dir, id)
		ioutil.WriteFile(calName, output, 0644)
		// date
		description := getDescription(day)
		dateName := fmt.Sprintf("%s/.tmp/date-%s.txt", Dir, id)
		ioutil.WriteFile(dateName, []byte(description), 0644)
		showCmd("figlet", strings.Trim(fmt.Sprintf("%v", day), "[]"))
	}
}

// generate new months
func genMonths(months [][2]int) {
	tex := fmt.Sprintf("%s/ifc.tex", Dir)
	pdf := fmt.Sprintf("%s/ifc.pdf", Dir)
	for _, month := range months {
		var idRaw string
		for i := 0; i < 2; i++ {
			idRaw += fmt.Sprintf("%02d-", month[i])
		}
		id := idRaw[:len(idRaw)-1]
		f := fmt.Sprintf("%s/.tmp/ifc-month-%s.pdf", Dir, id)
		fmt.Printf(f)
		if _, err := os.Stat(f); err == nil {
			continue
		}
		runCmd("sed", "-i", "22s/{.*}/{0}/", tex)
		for i := 0; i < 2; i++ {
			line := i + 23
			sedExp := fmt.Sprintf("%ds/{.*}/{%d}/", line, month[i])
			runCmd("sed", "-i", sedExp, tex)
		}
		runCmd("sed", "-i", "25s/{.*}/{}/", tex)
		runCmd("xelatex", tex)
		showCmd("figlet", strings.Trim(fmt.Sprintf("%v", month), "[]"))
		runCmd("cp", pdf, f)
	}
}

// generate new years
func genYears(years []int) {
	tex := fmt.Sprintf("%s/ifc.tex", Dir)
	pdf := fmt.Sprintf("%s/ifc.pdf", Dir)
	for _, year := range years {
		f := fmt.Sprintf("%s/.tmp/ifc-year-%d.pdf", Dir, year)
		if _, err := os.Stat(f); err == nil {
			continue
		}
		runCmd("sed", "-i", "22s/{.*}/{1}/", tex)
		sedExp := fmt.Sprintf("23s/{.*}/{%d}/", year)
		runCmd("sed", "-i", sedExp, tex)
		runCmd("sed", "-i", "24s/{.*}/{}/", tex)
		runCmd("sed", "-i", "25s/{.*}/{}/", tex)
		runCmd("xelatex", tex)
		showCmd("figlet", strings.Trim(fmt.Sprintf("%v", year), "[]"))
		runCmd("cp", pdf, f)
	}
}

// update website content
func updateContent(tzDates map[int][3]int, nearest Nearest) {
	seq := [3]string{"previous", "this", "next"}
	// copy all content based on timezone dates (locally)
	for tzShift, date := range tzDates {
		var idRaw, id, tzId string
		for i := 0; i < 3; i++ {
			idRaw += fmt.Sprintf("%02d-", date[i])
		}
		id = idRaw[:len(idRaw)-1]
		// copy day content
		calName := fmt.Sprintf("%s/.tmp/ifc-%s.png", Dir, id)
		dateName := fmt.Sprintf("%s/.tmp/date-%s.txt", Dir, id)
		calTzName := fmt.Sprintf("%s/cal/cal%+d.png", DirC, tzShift)
		dateTzName := fmt.Sprintf("%s/date/date%+d.txt", DirC, tzShift)
		runCmd("cp", calName, calTzName)
		runCmd("cp", dateName, dateTzName)
		// copy month content
		for i := -1; i < 2; i++ {
			year := nearest.Months[date][i][0]
			month := nearest.Months[date][i][1]
			id = fmt.Sprintf("month-%d-%02d", year, month)
			calName := fmt.Sprintf("%s/.tmp/ifc-%s.pdf", Dir, id)
			tzId = fmt.Sprintf("%s-month%+d", seq[i+1], tzShift)
			calTzName := fmt.Sprintf("%s/ifc-%s.pdf", DirD, tzId)
			runCmd("cp", calName, calTzName)
		// copy year content
		}
		for i := -1; i < 2; i++ {
			year := nearest.Years[date][i]
			id = fmt.Sprintf("year-%d", year)
			calName := fmt.Sprintf("%s/.tmp/ifc-%s.pdf", Dir, id)
			tzId = fmt.Sprintf("%s-year%+d", seq[i+1], tzShift)
			calTzName := fmt.Sprintf("%s/ifc-%s.pdf", DirD, tzId)
			runCmd("cp", calName, calTzName)
		}
	}
	// copy eternal content
	calEternal := fmt.Sprintf("%s/ifc-eternal.pdf", DirE)
	calEternalDownload := fmt.Sprintf("%s/ifc-eternal.pdf", DirD)
	runCmd("cp", calEternal, calEternalDownload)
	// remove all previous data
	runCmd("rm", "-rf", "/var/www/kalendaro")
	// copy all content (globally)
	runCmd("cp", "-r", "/root/kalendaro/website", "/var/www/kalendaro")
	// restart systemctl
	runCmd("systemctl", "restart", "nginx.service")
}

// creates temporary directories
func makeTmpDirs() {
	os.MkdirAll("/root/kalendaro/.tmp", os.ModePerm)
	os.MkdirAll("/root/kalendaro/website/content", os.ModePerm)
	os.MkdirAll("/root/kalendaro/website/content/cal", os.ModePerm)
	os.MkdirAll("/root/kalendaro/website/content/date", os.ModePerm)
	os.MkdirAll("/root/kalendaro/website/download", os.ModePerm)
}

func main() {
	// make all missing directories
	makeTmpDirs()
	// set date for every timezone (expressed in minutes)
	tzShifts := [...]int{-720, -660, -600, -570, -540, -480, -420, -360,
	-300, -240, -210, -180, -120, -60, 0, 60, 120, 180, 210, 240, 270, 300,
	330, 345, 360, 390, 420, 480, 525, 540, 570, 600, 630, 660, 720, 765,
	780, 840}
	tzDates := make(map[int][3]int, 38)
	setTzDates(tzShifts, tzDates)
	// get unique dates to generate
	uDaysGen := getUniqueDays(tzShifts, tzDates)
	// get unique months, years to generate
	nearest := getNearest(uDaysGen)
	gen := getGen(uDaysGen, nearest)
	uMonthsGen := getUniqueMonths(gen.Months)
	uYearsGen := getUniqueYears(gen.Years)
	// generate all content
	genDays(uDaysGen)
	genMonths(uMonthsGen)
	genYears(uYearsGen)
	// update all content
	updateContent(tzDates, nearest)
}
