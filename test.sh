#!/bin/bash

trap 'term' 15

term()
{
  echo "Caught Signal - sleep 5"
  sleep 5
  echo "Done."
  exit 1
}

X=0
while :
do
  echo "$X"
  X=`expr ${X} + 1`
  sleep 1
done