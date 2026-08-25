#!/bin/sh
echo PING | nc -w 3 127.0.0.1 3310 | grep -q PONG
