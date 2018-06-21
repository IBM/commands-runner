/*
###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
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
