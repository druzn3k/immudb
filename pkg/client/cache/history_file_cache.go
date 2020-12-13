/*
Copyright 2019-2020 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/golang/protobuf/proto"
)

type historyFileCache struct {
	dir string
}

// NewHistoryFileCache returns a new history file cache
func NewHistoryFileCache(dir string) HistoryCache {
	return &historyFileCache{dir: dir}
}

func (history *historyFileCache) Get(serverID string, dbName string) (*schema.ImmutableState, error) {
	statesDir := filepath.Join(history.dir, serverID)
	statesFileInfos, err := history.getStatesFileInfos(statesDir)
	if err != nil {
		return nil, err
	}

	if len(statesFileInfos) == 0 {
		return nil, nil
	}

	prevStateFileName := statesFileInfos[len(statesFileInfos)-1].Name()
	prevStateFilePath := filepath.Join(statesDir, prevStateFileName)
	return history.unmarshalRoot(prevStateFilePath, dbName)
}

func (history *historyFileCache) Walk(
	serverID string, databasename string,
	f func(*schema.ImmutableState) interface{},
) ([]interface{}, error) {
	statesDir := filepath.Join(history.dir, serverID)
	statesFileInfos, err := history.getStatesFileInfos(statesDir)
	if err != nil {
		return nil, err
	}

	if len(statesFileInfos) == 0 {
		return nil, nil
	}

	results := make([]interface{}, 0, len(statesFileInfos))

	for _, stateFileInfo := range statesFileInfos {
		stateFilePath := filepath.Join(statesDir, stateFileInfo.Name())
		state, err := history.unmarshalRoot(stateFilePath, databasename)
		if err != nil {
			return nil, err
		}
		results = append(results, f(state))
	}

	return results, nil
}

func (history *historyFileCache) Set(state *schema.ImmutableState, serverID string, dbName string) error {
	statesDir := filepath.Join(history.dir, serverID)
	if err := os.MkdirAll(statesDir, os.ModePerm); err != nil {
		return fmt.Errorf("error ensuring states dir %s exists: %v", statesDir, err)
	}
	stateFilePath := filepath.Join(statesDir, ".state")

	//at run first the file does not exist
	input, _ := ioutil.ReadFile(stateFilePath)

	lines := strings.Split(string(input), "\n")
	raw, err := proto.Marshal(state)
	if err != nil {
		return err
	}

	newState := dbName + ":" + base64.StdEncoding.EncodeToString(raw) + "\n"
	var exists bool
	for i, line := range lines {
		if strings.Contains(line, dbName+":") {
			exists = true
			lines[i] = newState
		}
	}
	if !exists {
		lines = append(lines, newState)
	}

	output := strings.Join(lines, "\n")

	if err = ioutil.WriteFile(stateFilePath, []byte(output), 0644); err != nil {
		return fmt.Errorf(
			"error writing state %d to file %s: %v",
			state.TxId, stateFilePath, err)
	}

	return nil
}

func (history *historyFileCache) getStatesFileInfos(dir string) ([]os.FileInfo, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error ensuring states dir %s exists: %v", dir, err)
	}

	statesFileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading states dir %s: %v", dir, err)
	}

	return statesFileInfos, nil
}

func (history *historyFileCache) unmarshalRoot(fpath string, dbName string) (*schema.ImmutableState, error) {
	state := &schema.ImmutableState{}
	raw, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("error reading state from %s: %v", fpath, err)
	}

	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		if strings.Contains(line, dbName+":") {
			r := strings.Split(line, ":")

			if len(r) != 2 {
				return nil, fmt.Errorf("could not find previous state")
			}

			oldRoot, err := base64.StdEncoding.DecodeString(r[1])
			if err != nil {
				return nil, fmt.Errorf("could not find previous state")
			}

			if err = proto.Unmarshal(oldRoot, state); err != nil {
				return nil, fmt.Errorf("error unmarshaling state from %s: %v", fpath, err)
			}
			return state, nil
		}
	}

	return nil, nil
}
