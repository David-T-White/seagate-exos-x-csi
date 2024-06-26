#!/bin/bash

function setNamespace()
{
    namespace=$CSI_NAMESPACE
    if [ -z "$namespace" ]
    then
        export CSI_NAMESPACE=csi
        namespace=$CSI_NAMESPACE
    fi
    echo "[] using namespace=$namespace"
}

function runCommand()
{
    echo ""
    echo "RUN>> $1"
    eval $1
}

function pause()
{
    echo ""
    read -s -n 1 -p "===== Press any key to contine ====="
    echo ""
}

function banner()
{
    echo ""
    echo "====================  $1  ===================="
    echo ""
}