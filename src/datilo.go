package main

import (
	"bytes"
	"fmt"
	"time"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// convenient date info struct
type DateInfo struct {
	Year      int
	Month     int
	MonthName string
	Day       int
}

// checks if year is leap
func ifLeapYear(year int) bool {
	leapYear := year%4 == 0 && !(year%100 == 0) || year%400 == 0
	return leapYear
}

// calculates year day according to Gregorian calendar
func calcYearDay(month int, monthDay int, leapYear bool) int {
	yearDay := 0
	monthDays := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if leapYear {
		monthDays[1] = 29
	}
	for _, monthDays := range monthDays[:month-1] {
		yearDay += monthDays
	}
	yearDay += monthDay
	return yearDay
}

// calculates date according to IFC
func calcDateIFC(day int, leapYear bool) (int, int) {
	const leapDay = 169
	yearDay := 365
	// calculate date
	month := (day / 28) + 1 // 1 2 ... 13
	monthDay := day % 28    // 1 2 ... 0 (<- 28 [will be reassigned])
	// recalculate date according to the leap day
	if leapYear {
		yearDay = 366
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

// returns IFC date as DateInfo struct
func getDateInfo(timezoneShiftMinutes int) DateInfo {
	const minutesPerHour = 60
	monthNamesEO := [13]string{"januaro", "februaro", "marto",
		"aprilo", "majo", "junio", "sunio", "julio", "aŭgusto",
		"septembro", "oktobro", "novembro", "decembro"}
	place := time.FixedZone("UTC", timezoneShiftMinutes*minutesPerHour)
	timestamp := time.Now().In(place)
	// get Gregorian date
	yearG, monthG, dayG := timestamp.Year(), int(timestamp.Month()), timestamp.Day()
	// calculate IFC date info
	leapYear := ifLeapYear(yearG)
	dayInYear := calcYearDay(monthG, dayG, leapYear)
	year := yearG
	month, day := calcDateIFC(dayInYear, leapYear)
	monthName := monthNamesEO[month-1]
	// package IFC date info
	dateInfo := DateInfo{
		Year:      year,
		Month:     month,
		MonthName: monthName,
		Day:       day,
	}
	return dateInfo
}

// wrapper for explicit command execution
func showCmd(app string, args ...string) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(app, args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = "/etc/ifc-website", &stdout, &stderr
	// print output of command
	cmd.Run()
	fmt.Println(stdout.String())
	cmd.Wait()
}

// wrapper for silent command execution
func runCmd(app string, args ...string) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(app, args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = "/etc/ifc-website", &stdout, &stderr
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
func setTzDates(timezoneShifts [38]int, timezoneDates map[int][3]int) {
	for _, timezoneShift := range timezoneShifts {
		dateInfo := getDateInfo(timezoneShift)
		timezoneDates[timezoneShift] = [3]int{dateInfo.Year, dateInfo.Month, dateInfo.Day}
	}
}

// get unique dates from timezone dates
func getUniqueDays(timezoneShifts [38]int, timezoneDates map[int][3]int) [][3]int {
	uniqueDaysBool := make(map[[3]int]bool, 3)
	var uniqueDays [][3]int
	for _, timezoneShift := range timezoneShifts {
		date := timezoneDates[timezoneShift]
		uniqueDaysBool[date] = true
	}
	for date, _ := range uniqueDaysBool {
		uniqueDays = append(uniqueDays, date)
	}
	fmt.Println(uniqueDays)
    	return uniqueDays
}

// get unique months
func getUniqueMonths(months [][2]int) [][2]int {
	uniqueMonthsBool := make(map[[2]int]bool)
	var uniqueMonths [][2]int
	for _, month := range months {
		uniqueMonthsBool[month] = true
	}
	for month, _ := range uniqueMonthsBool {
		uniqueMonths = append(uniqueMonths, month)
	}
	return uniqueMonths
}

// get unique years
func getUniqueYears(years []int) []int {
	uniqueYearsBool := make(map[int]bool)
	var uniqueYears []int
	for _, year := range years {
		uniqueYearsBool[year] = true
	}
	for year, _ := range uniqueYearsBool {
		uniqueYears = append(uniqueYears, year)
	}
	return uniqueYears
}

// get month and years to generate
func getMonthsYearsGen(uniqueDates [][3]int) ([][2]int, []int, map[[3]int]map[int][2]int, map[[3]int]map[int]int) {
	var monthsGen [][2]int
	var yearsGen []int
	// initialize outer maps
	months := make(map[[3]int]map[int][2]int)
	years := make(map[[3]int]map[int]int)
	// initialize inner maps
	for _, date := range uniqueDates {
    		months[date] = make(map[int][2]int)
    		years[date] = make(map[int]int)
	}
	// get previous, current, next months and years for every date
	for _, date := range uniqueDates {
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
		// append
		for i := -1; i < 2; i++ {
			monthsGen = append(monthsGen, months[date][i])
			yearsGen = append(yearsGen, years[date][i])
		}
	}
	return monthsGen, yearsGen, months, years
}

// get date description
func getDescription(dates [3]int) string {
	month, day := dates[1], dates[2]
	monthNamesIFC_EO := [13]string{"januaro", "februaro", "marto",
		"aprilo", "majo", "junio", "sunio", "julio", "aŭgusto",
		"septembro", "oktobro", "novembro", "decembro"}
	description := fmt.Sprintf("Hodiaŭ estas la %d de %s", day, monthNamesIFC_EO[month-1])
	return description
}

// generate new days
func genDays(days [][3]int) {
	for _, day := range days {
		if _, err := os.Stat(fmt.Sprintf("/etc/ifc-website/.tmp/ifc-%d-%02d-%02d.png", day[0], day[1], day[2])); err == nil {
			continue
	}
		runCmd("sed", "-i", "22s/{.*}/{0}/", "/etc/ifc-website/ifc.tex")
		for i := 0; i < 3; i++ {
			lineNum := i + 23
			runCmd("sed", "-i", strconv.Itoa(lineNum)+"s/{.*}/{"+strconv.Itoa(day[i])+"}/", "/etc/ifc-website/ifc.tex")
		}
		// calendar
		runCmd("xelatex", "/etc/ifc-website/ifc.tex")
		cmd := exec.Command("pdftoppm", "-png", "/etc/ifc-website/ifc.pdf")
		output, _ := cmd.Output()
		cal_name := fmt.Sprintf("/etc/ifc-website/.tmp/ifc-%d-%02d-%02d.png", day[0], day[1], day[2])
		ioutil.WriteFile(cal_name, output, 0644)
		// date
		description := getDescription(day)
		date_name := fmt.Sprintf("/etc/ifc-website/.tmp/date-%d-%02d-%02d.txt", day[0], day[1], day[2])
		ioutil.WriteFile(date_name, []byte(description), 0644)
		showCmd("figlet", strings.Trim(fmt.Sprintf("%v", day), "[]"))
	}
}

// generate new months
func genMonths(months [][2]int) {
	for _, month := range months {
		if _, err := os.Stat(fmt.Sprintf("/etc/ifc-website/.tmp/ifc-month-%d-%02d.pdf", month[0], month[1])); err == nil {
			continue
		}
		runCmd("sed", "-i", "22s/{.*}/{0}/", "/etc/ifc-website/ifc.tex")
		for i := 0; i < 2; i++ {
			lineNum := i + 23
			runCmd("sed", "-i", strconv.Itoa(lineNum)+"s/{.*}/{"+strconv.Itoa(month[i])+"}/", "/etc/ifc-website/ifc.tex")
		}
		runCmd("sed", "-i", "25s/{.*}/{}/", "/etc/ifc-website/ifc.tex")
		runCmd("xelatex", "/etc/ifc-website/ifc.tex")
		showCmd("figlet", strings.Trim(fmt.Sprintf("%v", month), "[]"))
		runCmd("cp", "/etc/ifc-website/ifc.pdf", fmt.Sprintf("/etc/ifc-website/.tmp/ifc-month-%d-%02d.pdf", month[0], month[1]))
	}
}

// generate new years
func genYears(years []int) {
	for _, year := range years {
		if _, err := os.Stat(fmt.Sprintf("/etc/ifc-website/.tmp/ifc-year-%d.pdf", year)); err == nil {
			continue
		}
		runCmd("sed", "-i", "22s/{.*}/{1}/", "/etc/ifc-website/ifc.tex")
		runCmd("sed", "-i", "23s/{.*}/{"+strconv.Itoa(year)+"}/", "/etc/ifc-website/ifc.tex")
		runCmd("sed", "-i", "24s/{.*}/{}/", "/etc/ifc-website/ifc.tex")
		runCmd("sed", "-i", "25s/{.*}/{}/", "/etc/ifc-website/ifc.tex")
		runCmd("xelatex", "/etc/ifc-website/ifc.tex")
		showCmd("figlet", strings.Trim(fmt.Sprintf("%v", year), "[]"))
		runCmd("cp", "/etc/ifc-website/ifc.pdf", fmt.Sprintf("/etc/ifc-website/.tmp/ifc-year-%d.pdf", year))
	}
}

// update website content
func updateContent(timezoneDates map[int][3]int, months map[[3]int]map[int][2]int, years map[[3]int]map[int]int) {
	sequence := [3]string{"previous", "this", "next"}
	// copy all content based on timezone dates (locally)
	for timezoneShift, date := range timezoneDates {
		// copy day content
		year, month, day := date[0], date[1], date[2]
		cal_name := fmt.Sprintf("/etc/ifc-website/.tmp/ifc-%d-%02d-%02d.png", year, month, day)
		date_name := fmt.Sprintf("/etc/ifc-website/.tmp/date-%d-%02d-%02d.txt", year, month, day)
		cal_name_tz := fmt.Sprintf("/etc/ifc-website/website/content/cal/cal%+d.png", timezoneShift)
		date_name_tz := fmt.Sprintf("/etc/ifc-website/website/content/date/date%+d.txt", timezoneShift)
		runCmd("cp", cal_name, cal_name_tz)
		runCmd("cp", date_name, date_name_tz)
		// copy month content
		for i := -1; i < 2; i++ {
			year, month = months[date][i][0], months[date][i][1]
			cal_name := fmt.Sprintf("/etc/ifc-website/.tmp/ifc-month-%d-%02d.pdf", year, month)
			cal_name_tz := fmt.Sprintf("/etc/ifc-website/website/download/ifc-%s-month%+d.pdf", sequence[i+1], timezoneShift)
			runCmd("cp", cal_name, cal_name_tz)
		// copy year content
		}
		for i := -1; i < 2; i++ {
			year = years[date][i]
			cal_name := fmt.Sprintf("/etc/ifc-website/.tmp/ifc-year-%d.pdf", year)
			cal_name_tz := fmt.Sprintf("/etc/ifc-website/website/download/ifc-%s-year%+d.pdf", sequence[i+1], timezoneShift)
			runCmd("cp", cal_name, cal_name_tz)
		}
	}
	// copy universal content
	cal_uni := "/etc/ifc-website/saved/ifc-eternal.pdf"
	cal_uni_web := "/etc/ifc-website/website/download/ifc-eternal.pdf"
	runCmd("cp", cal_uni, cal_uni_web)
	// remove all previous data
	runCmd("rm", "-rf", "/var/www/kalendaro")
	// copy all content (globally)
	runCmd("cp", "-r", "/etc/ifc-website/website", "/var/www/kalendaro")
	// restart systemctl
	runCmd("systemctl", "restart", "nginx.service")
}

// creates temporary directories
func makeTmpDirs() {
	os.MkdirAll("/etc/ifc-website/.tmp", os.ModePerm)
	os.MkdirAll("/etc/ifc-website/website/content", os.ModePerm)
	os.MkdirAll("/etc/ifc-website/website/content/cal", os.ModePerm)
	os.MkdirAll("/etc/ifc-website/website/content/date", os.ModePerm)
	os.MkdirAll("/etc/ifc-website/website/download", os.ModePerm)
}

func main() {
	// make all missing directories
	makeTmpDirs()
	// set date for every timezone (expressed in minutes)
	timezoneShifts := [...]int{-720, -660, -600, -570, -540, -480, -420, -360, -300,
	-240, -210, -180, -120, -60, 0, 60, 120, 180, 210, 240, 270, 300, 330, 345, 360,
	390, 420, 480, 525, 540, 570, 600, 630, 660, 720, 765, 780, 840}
	timezoneDates := make(map[int][3]int, 38)
	setTzDates(timezoneShifts, timezoneDates)
	// get unique dates to generate and calculate other stuff
	uniqueDaysGen := getUniqueDays(timezoneShifts, timezoneDates)
	// get unique months, years to generate
	monthsGen, yearsGen, months, years := getMonthsYearsGen(uniqueDaysGen)
	uniqueMonthsGen := getUniqueMonths(monthsGen)
	uniqueYearsGen := getUniqueYears(yearsGen)
	// generate all content
	genDays(uniqueDaysGen)
	genMonths(uniqueMonthsGen)
	genYears(uniqueYearsGen)
	// update all content
	updateContent(timezoneDates, months, years)
}
