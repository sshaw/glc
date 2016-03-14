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

const usage = `usage: glc [-a token] [-d path] [-e name...] [-i name...]
	   [-r name...] [-u name...] [-w seconds] command

Command must be one of:
  print		Print activity containing non-permanent links, include their permanent versions
  correct	Update the original event, replacing each non-permanent link with its permanent version
  comment	Create a comment on the event that includes the permanent version of each non-permanent link

Options:
  -a token          --auth=token		GitHub API token, all activity will be performed as the associated user
  -d path	    --db=path			Where to store the DB, defaults to $HOME/.glc/
  -e name[,name]    --event=name[,name...]	Only process the named GitHub events
  -i name[,name]    --ignore-files=name[,name]  Ignore links to these file basenames

  -r name[,name...] --repos=name[,name...]	Monitor the named repositories, name must be in user/repo format
  --include-repos=name[,name...]		name can also be a file with one repository per line
  --exclude-repos=name[,name...]		Do not monitor the named repositories, name must be in user/repo format
						name can also be a file with one repository per line

  -u name[,name...] --users=name[,name...]	Monitor repositories owned by the given usernames name can also be a file
  --include-users=name[,name...]		with one repository per line
  --exclude-users=name[,name...]		Do not monitor repositories owned by the given usernames name can also be a
						file with one repository per line
  -w seconds        --wait=seconds		Retrieve events every seconds seconds, defaults to 10
`

const (
	commandPrint = "print"
	commandCorrect = "correct"
	commandComment = "comment"

	defaultWait = 10
)

var (
	includeRepos string
	excludeRepos string
	includeUsers string
	excludeUsers string

	auth string
	daemonize bool
	db string
	events string
	ignoreFiles string
	runOnce bool
	wait int
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

	defer file.Close()

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
	fmt.Println(strings.Repeat("-", 18))
	// TODO: Sometimes should use "Number" other times "ID"
	fmt.Printf("%6s: %s\n%6s: %s\n%6s: %d\n", "Repo", event.Repo, "Event", event.Type, "Number", event.Number)
	fmt.Println(strings.Repeat("-", 18))

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

		time.Sleep(time.Duration(wait) * time.Second)
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
	flag.BoolVar(&daemonize, "b", false, "")
	flag.BoolVar(&daemonize, "background", false, "")
	flag.StringVar(&db, "d", "", "")
	flag.StringVar(&db, "db", "", "")
	flag.StringVar(&events, "e", "", "")
	flag.StringVar(&events, "events", "", "")
	flag.StringVar(&ignoreFiles, "i", "", "")
	flag.StringVar(&ignoreFiles, "ignore-files", "", "")
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
	flag.IntVar(&wait, "w", defaultWait, "")
	flag.IntVar(&wait, "wait", defaultWait, "")

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
	if wait < 0 {
		fail("wait must be > 0")
	}

	// GLCOptions
	if auth != "" {
		glcOptions.AccessToken = auth
	}

	if db != "" {
		glcOptions.DB = db
	}

	// EventOptions
	if ignoreFiles != "" {
		eventOptions.IgnoreFiles = parseNameArg(ignoreFiles)
	}

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
