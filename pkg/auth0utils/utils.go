package auth0utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/auth0.v1/management"
	"net/url"
	"sort"
)

type ClientList []*management.Client

func (l ClientList) AppID(name string) (string, error) {
	for _, a := range l {
		if *a.Name == name {
			return *a.ClientID, nil
		}
	}
	return "", fmt.Errorf("unable to find client id by name '%s'", name)
}

// SortUniq gets a slice of inerface and return sorted list of uniques elements
func SortUniq(s []interface{}) []interface{} {
	keys := make(map[interface{}]struct{})
	list := make([]interface{}, 0)
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}
	// Sort list without duplicates
	sort.Slice(list, func(i, j int) bool {
		return list[i].(string) < list[j].(string)
	})
	return list
}

// AddItem adds item to slice, sort slice and returns number of changed elements
func AddItem(s *[]interface{}, elem ...interface{}) (added int) {
	// Create a map with keys from slice to check if element in list already exists
	m := make(map[interface{}]struct{})
	// Populate map
	for _, i := range *s {
		m[i] = struct{}{}
	}
	//create a list of valid urls
	var list []interface{}
	for _, i := range elem {
		u, parseErr := url.ParseRequestURI(i.(string))
		if parseErr != nil {
			log.Warningf("URL is invalid, %s", parseErr)
		} else {
			u := u.String()
			if _, ok := m[u]; !ok {
				list = append(list, u)
				added++
			}
		}
	}
	log.Debug("added URLs:", elem)
	*s = SortUniq(append(*s, list...))
	return
}

// DeleteItem deletes string item from slice
func DeleteItem(s *[]interface{}, elem ...interface{}) (deleted int) {
	m := make(map[interface{}]struct{})
	for _, i := range *s {
		m[i] = struct{}{}
	}
	// delete elements from map
	for _, i := range elem {
		if _, ok := m[i]; ok {
			log.Debug("Deleting url:", elem)
			delete(m, i)
			deleted++
		}
	}
	// iterate over map and fill new list
	resultSlice := make([]interface{}, 0)
	for k := range m {
		resultSlice = append(resultSlice, k)
	}
	// Replace incoming slice with newly generated
	log.Debug("deleted URLs:", elem)
	*s = SortUniq(resultSlice)
	return
}
