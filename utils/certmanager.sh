#!/bin/bash

cmd_exists (){
    type "$1" &> /dev/null;
}

for sudo_cmd in "kdesu" "gksu" "pkexec"; do
   if cmd_exists $sudo_cmd ; then
      su_graph=$sudo_cmd
      echo $su_graph
      break
   fi
done

export PATH=$PATH:/sbin/
if cmd_exists update-ca-certificates; then
    ca_path=/usr/local/share/ca-certificates/
    ca_update_cmd=update-ca-certificates
else
if cmd_exists update-ca-trust; then
    ca_path=/usr/share/ca-certificates/trust-source/anchors/
    ca_update_cmd=update-ca-trust
else
    $su_graph apt-get install ca-certificates
    ca_path=/usr/local/share/ca-certificates/
    ca_update_cmd=update-ca-certificates
fi
fi

$su_graph cp $1 $ca_path/createAgentLocal.crt
$su_graph $ca_update_cmd
#Alway run install, it does not hurt

if cmd_exists apt-get; then
    DBDIR="$HOME/.pki/nssdb"
    $su_graph apt-get install libnss3-tools
    
    # add the cert only if the db exists already
    stat $DBDIR
    if [ "$?" -eq "0" ]; then
        certutil -d sql:$DBDIR -A -t "C,," -n Arduino -i $1
    fi

fi
exit $?
