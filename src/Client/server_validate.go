package main

import "regexp"

var serverRegExp = regexp.MustCompile(`^\w*\.\w*\.\w*$`)

func serverNameValid(serverName string) bool {

	return serverRegExp.MatchString(serverName)
}
