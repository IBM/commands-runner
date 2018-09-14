#!/bin/sh

###############################################################################
# Licensed Materials - Property of IBM Copyright IBM Corporation 2017, 2018. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP
# Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################

echo "Start Step1: $1"
sleep 1
if [ -f config.yml ]; then
  cat config.yml
fi
echo "End: $1"
exit 0
