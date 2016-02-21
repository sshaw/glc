package main

import (
	"bufio"
	"fmt"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/sshaw/glc"
)

// TODO: --ignore-files=xxx
const usage = `usage: glc [-di seconds] [-a token] [-u name...] [-r name...] command
Command must be one of:
  print		Print activity containing non-permanent links, include their permanent versions
  correct	Update the original event, replacing each non-permanent link with its permanent version
  comment	Create a comment on the event that includes the permanent version of each non-permanent link

Options:
  -a token          --auth=token		GitHub API token, comments/suggestions will be performed as the associated user
  -b	   	    --background		Run in the background as a daemon
  -d path	    --db=path			Where to store the DB, defaults to $HOME/.glc/
  -e name[,name]    --event=name[,name...]	Only process the named GitHub events
  -i seconds        --interval=seconds		Retrieve events every seconds seconds, defaults to 5

  -r name[,name...] --repos=name[,name...]	Monitor the named repositories, name must be in user/repo format
  --include-repos=name[,name...]		name can also be a file with one repository per line
  --exclude-repos=name[,name...]		Do not monitor the named repositories, name must be in user/repo format
						name can also be a file with one repository per line

  -u name[,name...] --users=name[,name...]	Monitor repositories owned by the given usernames name can also be a file
  --include-users=name[,name...]		with one repository per line
  --exclude-users=name[,name...]		Do not monitor repositories owned by the given usernames name can also be a
						file with one repository per line
`

const (
	commandPrint = "print"
	commandCorrect = "correct"
	commandComment = "comment"

	defaultInterval = 5
)

var (
	events string
	includeRepos string
	excludeRepos string
	includeUsers string
	excludeUsers string

	runOnce bool
	auth string
	daemonize bool
	db string
	interval int

	ignoreFiles = []string{
		"AUTHORS", "AUTHORS.txt", "AUTHORS.md", "AUTHORS.markdown",
		"CONTRIBUTING", "CONTRIBUTING.txt", "CONTRIBUTING.md", "CONTRIBUTING.markdown",
		"LICENSE", "LICENSE.txt",
		"ISSUE_TEMPLATE", "ISSUE_TEMPLATE.md", "ISSUE_TEMPLATE.markdown",
		"PULL_REQUEST_TEMPLATE", "PULL_REQUEST_TEMPLATE.md", "PULL_REQUEST_TEMPLATE.markdown",
	}
)

func printError(err string, args ...interface{}) {
	err = fmt.Sprintf("glc: %s\n", err)
	fmt.Fprintf(os.Stderr, err, args...)
}

func fail(err string, args ...interface{})  {
	printError(err, args...)
	os.Exit(1)
}

func eachLine(path string, cb func(string))  {
	file, err := os.Open(path)
	if err != nil {
		fail("failed to open file %s: %s", path, err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cb(scanner.Text())
	}

	if scanner.Err() != nil {
		fail("error reading file %s: %s", path, scanner.Err())
	}
}

func parseNameArg(arg string) []string {
	var names []string

	stat, err := os.Stat(arg)
	if err == nil {
		if !stat.Mode().IsRegular() {
			fail("not a file %s", arg)
		}

		eachLine(arg, func (name string) {
			names = append(names, name)
		})
	} else {
		// Assume it's not a file
		names = strings.Split(arg, ",")
	}

	return names
}

func commentOnEvent(event *glc.Event)  {
	fmt.Printf("Commenting on %s #%d by %s\n", event.Type, event.Number, event.Actor)

	id, err := event.Comment()
	if err != nil {
		printError("Comment failed: %s", err)
	} else {
		fmt.Printf("Comment successful, id=%d\n", id)
	}
}

func correctEvent(event *glc.Event)  {
	fmt.Printf("Correcting %s #%d by %s\n", event.Type, event.Number, event.Actor)

	err := event.Correct()
	if err != nil {
		printError("Correction failed: %s", err)
	} else {
		fmt.Println("Correction successful")
	}
}

func printEvent(event *glc.Event)  {
	fmt.Println(strings.Repeat("-", 17))
	// TODO: Sometimes should use "Number" other times "ID"
	fmt.Printf("%5s: %s\n%5s: %d\n", "Event", event.Type, "Number", event.Number)
	fmt.Println(strings.Repeat("-", 17))

	for i, correction := range(event.Corrections) {
		fmt.Printf("%2d. %-11s %s\n", i + 1, "Current:", correction.OldURL.String())
		fmt.Printf("%2s  %-11s %s\n", "", "Corrected:", correction.NewURL.String())
		fmt.Printf("%2s  %-11s %s\n\n", "", "Context:", correction.Context)
	}
}

func executeCommand(command string, glcOptions *glc.GLCOptions, eventOptions *glc.EventOptions) {
	checker := glc.New(glcOptions)
	for {
		events, err := checker.FindEvents(eventOptions)
		if err != nil {
			// TODO: if daemon just warn, don't exit
			fail("error finding events: %s", err)
		}

		for _, event := range(events) {
			switch command {
			case commandPrint:
				printEvent(event)
			case commandCorrect:
				correctEvent(event)
			case commandComment:
				commentOnEvent(event)
			}
		}

		if runOnce {
			break
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func main() {
	var glcOptions glc.GLCOptions
	var eventOptions glc.EventOptions

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	flag.StringVar(&auth, "a", "", "")
	flag.StringVar(&auth, "auth", os.Getenv("GLC_AUTH_TOKEN"), "")
	flag.BoolVar(&daemonize, "d", false, "")
	flag.BoolVar(&daemonize, "daemon", false, "")
	flag.StringVar(&events, "e", "", "")
	flag.StringVar(&events, "events", "", "")
	flag.IntVar(&interval, "i", defaultInterval, "")
	flag.IntVar(&interval, "interval", defaultInterval, "")
	flag.StringVar(&includeRepos, "r", "", "")
	flag.StringVar(&includeRepos, "repos", "", "")
	flag.BoolVar(&runOnce, "1", false, "")
	flag.BoolVar(&runOnce, "once", false, "")
	flag.StringVar(&includeRepos, "include-repos", "", "")
	flag.StringVar(&excludeRepos, "exclude-repos", "", "")
	flag.StringVar(&includeUsers, "u", "", "")
	flag.StringVar(&includeUsers, "users", "", "")
	flag.StringVar(&includeUsers, "include-users", "", "")
	flag.StringVar(&excludeUsers, "exclude-users", "", "")

	flag.Parse()

	command := flag.Arg(0)
	if command == "" {
		flag.Usage()
	}

	if command != commandPrint && command != commandCorrect && command != commandComment {
		printError("unknown command '%s'\n", command)
		flag.Usage()
	}

	if auth == "" && (command == commandCorrect || command == commandComment) {
		fail("use of the '%s' command requires a token", command)
	}

	// Process options
	if interval < 0 {
		fail("interval must be > 0")
	}

	// GLCOptions
	if auth != "" {
		glcOptions.AccessToken = auth
	}

	if db != "" {
		glcOptions.DB = db
	}

	// EventOptions
	if includeRepos != "" {
		eventOptions.IncludeRepos = parseNameArg(includeRepos)
	}

	if excludeRepos != "" {
		eventOptions.ExcludeRepos = parseNameArg(excludeRepos)
	}

	if includeUsers != "" {
		eventOptions.IncludeUsers = parseNameArg(includeUsers)
	}

	if excludeUsers != "" {
		eventOptions.ExcludeUsers = parseNameArg(excludeUsers)
	}

	executeCommand(command, &glcOptions, &eventOptions)
}
