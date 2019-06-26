/*
################################################################################
# Copyright 2019 IBM Corp. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################
*/
package clientManager

import (
	"crypto/sha512"
	"encoding/base64"
	"io/ioutil"
	"strconv"
	"time"
)

//NewToken creates Token base on time
func NewToken() string {
	time := strconv.FormatInt(time.Now().Unix(), 10)
	hasher := sha512.New()
	hasher.Write([]byte(time))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

//NewTokenFile creates a file containing a token
func NewTokenFile(filePath string) error {
	token := NewToken()
	return ioutil.WriteFile(filePath, []byte(token), 0644)
}
