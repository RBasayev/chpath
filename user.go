package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

type id struct {
	Id   int
	Name string
}

func getUID() id {
	var result id
	var username string
	var userObj *user.User

	username = strings.TrimSpace(owner)

	_, err := strconv.Atoi(username)
	if err != nil {
		// input is not numeric
		userObj, err = user.Lookup(username)
		if err != nil {
			fmt.Println("The user -", username, "- could not be found.")
			os.Exit(51)
		}
	} else {
		// input appears to be numeric
		userObj, err = user.LookupId(username)
		if err != nil {
			fmt.Println("The user ID", username, "could not be found.")
			os.Exit(52)
		}
	}

	result.Id, err = strconv.Atoi(userObj.Uid)
	check(err)
	result.Name = userObj.Username
	return result

}

func getGID() id {
	var result id
	var groupname string
	var userObj *user.Group

	groupname = strings.TrimSpace(group)

	_, err := strconv.Atoi(groupname)
	if err != nil {
		// input is not numeric
		userObj, err = user.LookupGroup(groupname)
		if err != nil {
			fmt.Println("The group -", groupname, "- could not be found.")
			os.Exit(53)
		}
	} else {
		// input appears to be numeric
		userObj, err = user.LookupGroupId(groupname)
		if err != nil {
			fmt.Println("The group ID", groupname, "could not be found.")
			os.Exit(54)
		}
	}

	result.Id, err = strconv.Atoi(userObj.Gid)
	check(err)
	result.Name = userObj.Name
	return result
}

func chUsers(uid id, gid id) {
	fnChown := func(thisCrumb crumb) {
		err := os.Chown(thisCrumb.PathActual, uid.Id, gid.Id)
		check(err)
		fmt.Printf("User: %s (%d) and Group: %s (%d) set for %s\n", uid.Name, uid.Id, gid.Name, gid.Id, thisCrumb.PathActual)
	}

	pathL2R(fnChown)

}
