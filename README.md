# GLC: The GitHub Link Checker

Monitor GitHub activity for [links that aren't permanent](https://help.github.com/articles/getting-permanent-links-to-files/), and do something.

**WIP**

## Install

	go get github.com/sshaw/glc/...

Put `$GOPATH/bin` (assuming `GOPATH` has one path) in your `PATH`.

## Usage

    usage: glc  [-a token] [-d path] [-e name...] [-i name...]
    			[-r name...] [-u name...] [-w seconds] command
    Command must be one of:
      print		Print activity containing non-permanent links, include their permanent versions
      correct	Update the original event, replacing each non-permanent link with its permanent version
      comment	Create a comment on the event that includes the permanent version of each non-permanent link

    Options:
      -a token          --auth=token			   GitHub API token, all activity will be performed as the associated user
      -d path	        --db=path				   Where to store the DB, defaults to $HOME/.glc/
      -e name[,name]    --event=name[,name...]	   Only process the named GitHub events
      -i name[,name]    --ignore-files=name[,name] Ignore links to these file's basenames

	  -r name[,name...] --repos=name[,name...]	   Monitor the named repositories, name must be in user/repo format
      --include-repos=name[,name...]			   name can also be a file with one repository per line
      --exclude-repos=name[,name...]			   Do not monitor the named repositories, name must be in user/repo format
	  											   name can also be a file with one repository per line

      -u name[,name...] --users=name[,name...]	   Monitor repositories owned by the given usernames name can also be a file
      --include-users=name[,name...]			   with one repository per line
      --exclude-users=name[,name...]			   Do not monitor repositories owned by the given usernames name can also be a
      	  										   file with one repository per line
      -w seconds        --wait=seconds			   Retrieve events every seconds seconds, defaults to 10

GLC monitors [public GitHub events](https://developer.github.com/v3/activity/events/#list-public-events), polling for new events every `-w` seconds (default `10`).
Currently only these events are monitored: `IssueEvent`, `IssueCommentEvent`, `PullRequestEvent`.

If no token is specified then your IP will be substantially [rate limited by GitHub](https://developer.github.com/v3/#rate-limiting). A token can be provided by the `-a` option
or by setting the `GLC_AUTH_TOKEN` environment variable.

The `correct` and `comment` options require a token.

GLC will not detect problematic links in older events unless they're still in GitHub's event feed and the event has not yet been seen.

### Examples

Only correct links in the repositories listed in `repos.txt`:

	glc -a TOKEN -r repos.txt correct

Only leave comments on new issues that weren't created by users `foo` and `bar`:

	glc -a TOKEN -e IssueEvent --exclude-users=foo,bar comment


## TODO

*  GitHub hooks

## Author

Skye Shaw [skye.shaw AT gmail.com]

## License

Released under the MIT License: http://www.opensource.org/licenses/MIT
