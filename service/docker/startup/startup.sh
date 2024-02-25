#!/usr/bin/env bash

# saner programming env: these switches turn some bugs into errors
set -o pipefail -o nounset

function start_ssh() {
    sudo service ssh start >/dev/null 2>&1
}

function set_gpios {
    local output_gpios=("10" "22" "27")
    local input_gpios=("4")
    for gpio in "${output_gpios[@]}"; do
        echo ${gpio} > /sys/class/gpio/export
        echo out > /sys/class/gpio/gpio${gpio}/direction
        echo ${gpio} > /sys/class/gpio/unexport
    done
    for gpio in "${input_gpios[@]}"; do
        echo ${gpio} > /sys/class/gpio/export
        echo in > /sys/class/gpio/gpio${gpio}/direction
        echo ${gpio} > /sys/class/gpio/unexport
    done
}

###
# This is the script's entry point, just like in any other programming language.
###
function main {
    # start_ssh
    set_gpios
    ldconfig 2>/dev/null
    exec "$@"
}

# Call main and don't do anything else
# It will pass the correct exit code to the OS
main "$@"
